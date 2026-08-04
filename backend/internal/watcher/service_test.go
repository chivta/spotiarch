package watcher

import (
	"context"
	"errors"
	"testing"

	"github.com/chivta/spotiarch/internal/shared/domain"
	"github.com/chivta/spotiarch/internal/spotify"
)

type addCall struct {
	playlistID string
	uris       []string
}

type fakeSpotify struct {
	playlist     *spotify.PlaylistResponse
	playlistErr  error
	items        []domain.SourceItem
	localFiles   int
	itemsErr     error
	itemsCalls   int
	addCalls     []addCall
	createdNames []string
	nextID       int
}

func (f *fakeSpotify) GetPlaylist(context.Context, string) (*spotify.PlaylistResponse, error) {
	return f.playlist, f.playlistErr
}

func (f *fakeSpotify) GetPlaylistItems(context.Context, string) ([]domain.SourceItem, int, error) {
	f.itemsCalls++
	return f.items, f.localFiles, f.itemsErr
}

func (f *fakeSpotify) CreateArchivePlaylist(_ context.Context, name, _ string) (string, error) {
	f.createdNames = append(f.createdNames, name)
	f.nextID++
	return "archive-playlist", nil
}

func (f *fakeSpotify) AddItems(_ context.Context, playlistID string, uris []string) error {
	f.addCalls = append(f.addCalls, addCall{playlistID, append([]string(nil), uris...)})
	return nil
}

type fakeWatchRepo struct {
	snapshotID     string
	localFileCount int
	updateCalls    int
}

func (f *fakeWatchRepo) UpdateSnapshot(_ context.Context, _ int, snapshotID string, localFileCount int) error {
	f.snapshotID = snapshotID
	f.localFileCount = localFileCount
	f.updateCalls++
	return nil
}

type fakeArchiveRepo struct {
	seen         []domain.SourceItem
	presentISRCs []string
	removed      int64
	unarchived   []domain.ArchiveTrack
	markedIDs    []int
	parts        []domain.ArchivePart
	created      []domain.ArchivePart
	countUpdates map[int]int
}

func (f *fakeArchiveRepo) UpsertSeen(_ context.Context, _ int, items []domain.SourceItem) error {
	f.seen = items
	return nil
}

func (f *fakeArchiveRepo) MarkRemoved(_ context.Context, _ int, presentISRCs []string) (int64, error) {
	f.presentISRCs = presentISRCs
	return f.removed, nil
}

func (f *fakeArchiveRepo) ListUnarchived(context.Context, int) ([]domain.ArchiveTrack, error) {
	return f.unarchived, nil
}

func (f *fakeArchiveRepo) MarkArchived(_ context.Context, ids []int) error {
	f.markedIDs = append(f.markedIDs, ids...)
	return nil
}

func (f *fakeArchiveRepo) CreatePart(_ context.Context, part *domain.ArchivePart) (int, error) {
	f.created = append(f.created, *part)
	return 100 + len(f.created), nil
}

func (f *fakeArchiveRepo) ListParts(context.Context, int) ([]domain.ArchivePart, error) {
	return f.parts, nil
}

func (f *fakeArchiveRepo) SetPartTrackCount(_ context.Context, partID, trackCount int) error {
	if f.countUpdates == nil {
		f.countUpdates = map[int]int{}
	}
	f.countUpdates[partID] = trackCount
	return nil
}

func newFixture() (*Service, *fakeSpotify, *fakeWatchRepo, *fakeArchiveRepo) {
	sp := &fakeSpotify{playlist: &spotify.PlaylistResponse{ID: "src", Name: "Source", SnapshotID: "snap-2"}}
	wr := &fakeWatchRepo{}
	ar := &fakeArchiveRepo{}
	return NewService(sp, wr, ar), sp, wr, ar
}

// The snapshot check is the whole point of polling: an unchanged playlist must
// not cost a full item fetch.
func TestProcessWatchSkipsUnchangedSnapshot(t *testing.T) {
	svc, sp, wr, ar := newFixture()
	sp.playlist.SnapshotID = "snap-1"

	if err := svc.ProcessWatch(context.Background(), domain.Watch{ID: 1, LastSnapshotID: "snap-1"}); err != nil {
		t.Fatalf("ProcessWatch: %v", err)
	}

	if sp.itemsCalls != 0 {
		t.Errorf("fetched items %d times for an unchanged snapshot, want 0", sp.itemsCalls)
	}
	if wr.updateCalls != 0 {
		t.Errorf("updated the snapshot %d times, want 0", wr.updateCalls)
	}
	if ar.seen != nil {
		t.Error("upserted tracks for an unchanged snapshot")
	}
}

func TestProcessWatchDiffsAndArchives(t *testing.T) {
	svc, sp, wr, ar := newFixture()
	sp.items = []domain.SourceItem{
		{URI: "spotify:track:a", ISRC: "ISRC-A"},
		{URI: "spotify:track:b", ISRC: "ISRC-B"},
	}
	sp.localFiles = 3
	ar.removed = 2
	ar.unarchived = []domain.ArchiveTrack{
		{ID: 11, URI: "spotify:track:a"},
		{ID: 12, URI: "spotify:track:b"},
	}

	if err := svc.ProcessWatch(context.Background(), domain.Watch{ID: 1, LastSnapshotID: "snap-1"}); err != nil {
		t.Fatalf("ProcessWatch: %v", err)
	}

	if len(ar.seen) != 2 {
		t.Errorf("upserted %d tracks, want 2", len(ar.seen))
	}
	// removal is decided by ISRC, not track id, because the same recording has
	// different ids across markets
	want := []string{"ISRC-A", "ISRC-B"}
	if len(ar.presentISRCs) != len(want) {
		t.Fatalf("passed %d present ISRCs, want %d", len(ar.presentISRCs), len(want))
	}
	for i, w := range want {
		if ar.presentISRCs[i] != w {
			t.Errorf("present ISRC %d = %q, want %q", i, ar.presentISRCs[i], w)
		}
	}

	if len(sp.addCalls) != 1 {
		t.Fatalf("made %d AddItems calls, want 1", len(sp.addCalls))
	}
	if len(sp.addCalls[0].uris) != 2 {
		t.Errorf("appended %d uris, want 2", len(sp.addCalls[0].uris))
	}
	if len(ar.markedIDs) != 2 {
		t.Errorf("marked %d tracks archived, want 2", len(ar.markedIDs))
	}
	if wr.snapshotID != "snap-2" {
		t.Errorf("stored snapshot %q, want %q", wr.snapshotID, "snap-2")
	}
	// local files cannot be archived, so the count is surfaced rather than dropped
	if wr.localFileCount != 3 {
		t.Errorf("stored local file count %d, want 3", wr.localFileCount)
	}
}

func TestProcessWatchCreatesFirstPartWhenNoneExists(t *testing.T) {
	svc, sp, _, ar := newFixture()
	sp.items = []domain.SourceItem{{URI: "spotify:track:a", ISRC: "ISRC-A"}}
	ar.unarchived = []domain.ArchiveTrack{{ID: 1, URI: "spotify:track:a"}}

	if err := svc.ProcessWatch(context.Background(), domain.Watch{ID: 1, LastSnapshotID: "old"}); err != nil {
		t.Fatalf("ProcessWatch: %v", err)
	}

	if len(ar.created) != 1 {
		t.Fatalf("created %d parts, want 1", len(ar.created))
	}
	if ar.created[0].PartNumber != 1 {
		t.Errorf("first part numbered %d, want 1", ar.created[0].PartNumber)
	}
}

// A playlist caps out near 10k tracks, so the archive rolls over into a numbered
// continuation playlist.
func TestProcessWatchRollsOverAtPartCap(t *testing.T) {
	svc, sp, _, ar := newFixture()
	ar.parts = []domain.ArchivePart{
		{ID: 7, WatchID: 1, PlaylistID: "part-1", PartNumber: 1, TrackCount: domain.ArchivePartCap - 1},
	}
	sp.items = []domain.SourceItem{{URI: "spotify:track:a", ISRC: "ISRC-A"}}
	ar.unarchived = []domain.ArchiveTrack{
		{ID: 1, URI: "spotify:track:a"},
		{ID: 2, URI: "spotify:track:b"},
		{ID: 3, URI: "spotify:track:c"},
	}

	if err := svc.ProcessWatch(context.Background(), domain.Watch{ID: 1, LastSnapshotID: "old"}); err != nil {
		t.Fatalf("ProcessWatch: %v", err)
	}

	if len(ar.created) != 1 {
		t.Fatalf("created %d continuation parts, want 1", len(ar.created))
	}
	if ar.created[0].PartNumber != 2 {
		t.Errorf("continuation numbered %d, want 2", ar.created[0].PartNumber)
	}

	if len(sp.addCalls) != 2 {
		t.Fatalf("made %d AddItems calls, want 2 (one per part)", len(sp.addCalls))
	}
	// exactly one track fits in the old part before it hits the cap
	if got := len(sp.addCalls[0].uris); got != 1 {
		t.Errorf("appended %d uris to the full part, want 1", got)
	}
	if sp.addCalls[0].playlistID != "part-1" {
		t.Errorf("first append went to %q, want %q", sp.addCalls[0].playlistID, "part-1")
	}
	if got := len(sp.addCalls[1].uris); got != 2 {
		t.Errorf("appended %d uris to the continuation, want 2", got)
	}
	if ar.countUpdates[7] != domain.ArchivePartCap {
		t.Errorf("part 1 count = %d, want %d", ar.countUpdates[7], domain.ArchivePartCap)
	}
	if len(ar.markedIDs) != 3 {
		t.Errorf("marked %d tracks archived, want 3", len(ar.markedIDs))
	}
}

// A failed poll must not advance the snapshot, or the change would be skipped
// forever on the next tick.
func TestProcessWatchKeepsSnapshotOnSpotifyError(t *testing.T) {
	svc, sp, wr, _ := newFixture()
	sp.itemsErr = &spotify.APIError{Status: 429}

	err := svc.ProcessWatch(context.Background(), domain.Watch{ID: 1, LastSnapshotID: "snap-1"})
	if !errors.Is(err, domain.ErrTooManyRequests) {
		t.Errorf("got error %v, want %v", err, domain.ErrTooManyRequests)
	}
	if wr.updateCalls != 0 {
		t.Errorf("advanced the snapshot %d times despite the error, want 0", wr.updateCalls)
	}
}

func TestProcessWatchEmptySourceMarksEverythingRemoved(t *testing.T) {
	svc, _, wr, ar := newFixture()
	ar.removed = 5

	if err := svc.ProcessWatch(context.Background(), domain.Watch{ID: 1, LastSnapshotID: "old"}); err != nil {
		t.Fatalf("ProcessWatch: %v", err)
	}

	if ar.presentISRCs == nil {
		t.Fatal("passed a nil ISRC slice; an empty source must still be diffed")
	}
	if len(ar.presentISRCs) != 0 {
		t.Errorf("passed %d present ISRCs, want 0", len(ar.presentISRCs))
	}
	if wr.snapshotID != "snap-2" {
		t.Errorf("stored snapshot %q, want %q", wr.snapshotID, "snap-2")
	}
}
