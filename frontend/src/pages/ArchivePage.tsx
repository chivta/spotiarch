import { useCallback, useEffect, useState } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";
import { deleteArchiveTrack, deleteWatch, getArchiveTracks, getWatch } from "../api";
import ErrorNotice from "../components/ErrorNotice";
import Layout from "../components/Layout";
import PlaylistPreviewCard from "../components/PlaylistPreviewCard";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import type { ArchiveTracksPage, WatchResponse } from "../types/models";
import { cardStyle, COLORS, CONTENT_WIDTH, primaryButtonStyle, secondaryButtonStyle } from "../ui";

const PAGE_SIZE = 50;

export default function ArchivePage() {
  const { id: idParam } = useParams();
  const id = Number(idParam);
  const [watch, setWatch] = useState<WatchResponse | null>(null);
  const [page, setPage] = useState<ArchiveTracksPage | null>(null);
  const [removedOnly, setRemovedOnly] = useState(false);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<unknown>(null);
  const [deletingURI, setDeletingURI] = useState<string | null>(null);
  const { lang } = useLanguage();
  const tx = translations[lang];
  const locale = lang === "uk" ? "uk-UA" : "en-US";
  const location = useLocation();
  const navigate = useNavigate();
  const verified = Boolean((location.state as { verified?: boolean } | null)?.verified);

  const load = useCallback(async () => {
    if (!Number.isInteger(id) || id <= 0) {
      navigate("/dashboard", { replace: true });
      return;
    }
    setError(null);
    try {
      const [nextWatch, nextPage] = await Promise.all([
        getWatch(id),
        getArchiveTracks(id, offset, PAGE_SIZE, removedOnly),
      ]);
      setWatch(nextWatch);
      setPage(nextPage);
    } catch (nextError) { setError(nextError); }
  }, [id, navigate, offset, removedOnly]);

  useEffect(() => { void load(); }, [load]);

  const handleDeleteTrack = async (uri: string) => {
    if (!window.confirm(tx.confirmDeleteTrack)) return;
    setDeletingURI(uri);
    setError(null);
    try { await deleteArchiveTrack(id, uri); await load(); }
    catch (nextError) { setError(nextError); }
    finally { setDeletingURI(null); }
  };

  const handleDeleteWatch = async () => {
    if (!window.confirm(tx.confirmDeleteWatch)) return;
    try { await deleteWatch(id); navigate("/dashboard", { replace: true }); }
    catch (nextError) { setError(nextError); }
  };

  const date = (value: string) => new Intl.DateTimeFormat(locale, { dateStyle: "medium" }).format(new Date(value));
  const pageFrom = page && page.total > 0 ? page.offset + 1 : 0;
  const pageTo = page ? Math.min(page.offset + page.tracks.length, page.total) : 0;

  return (
    <Layout>
      <div style={{ maxWidth: CONTENT_WIDTH, margin: "0 auto", padding: "48px 18px" }}>
        <Link to="/dashboard" style={{ color: COLORS.muted, fontSize: 13 }}>{tx.back}</Link>
        {verified && <div role="status" style={{ margin: "18px 0", padding: 14, borderRadius: 12, color: "#aef7c5", border: "1px solid rgba(30,215,96,0.35)", background: "rgba(30,215,96,0.08)" }}>{tx.tokenRemovable}</div>}
        {Boolean(error) && <div style={{ margin: "18px 0" }}><ErrorNotice error={error} /></div>}
        {!watch && !error && <div style={{ marginTop: 18, color: COLORS.muted }}>{tx.loading}</div>}
        {watch && (
          <>
            <div style={{ display: "flex", flexWrap: "wrap", alignItems: "end", justifyContent: "space-between", gap: 15, margin: "22px 0" }}>
              <h1 style={{ margin: 0, fontSize: "clamp(2.2rem, 5vw, 3.8rem)", letterSpacing: "-0.055em" }}>{tx.archiveTitle}</h1>
              <button onClick={handleDeleteWatch} style={{ ...secondaryButtonStyle, color: "#ff9e9e", borderColor: "rgba(255,107,107,0.35)" }}>{tx.deleteWatch}</button>
            </div>
            <PlaylistPreviewCard playlist={watch.source} />

            <section style={{ ...cardStyle, padding: "clamp(18px, 4vw, 28px)", marginTop: 18 }}>
              <h2 style={{ margin: "0 0 8px" }}>{tx.archiveRulesTitle}</h2>
              <p style={{ margin: "0 0 9px", color: COLORS.muted, lineHeight: 1.55 }}>{tx.archiveAppendOnly}</p>
              <p style={{ margin: 0, color: COLORS.muted, lineHeight: 1.55 }}>{tx.archiveManagedHere}</p>
              {watch.local_file_count > 0 && <div style={{ marginTop: 17, padding: 13, borderRadius: 11, color: "#ffe29a", background: "rgba(242,201,76,0.09)", border: "1px solid rgba(242,201,76,0.25)" }}>{tx.localFilesNotice(watch.local_file_count)}</div>}
            </section>

            <section style={{ ...cardStyle, padding: "clamp(18px, 4vw, 28px)", marginTop: 18 }}>
              <h2 style={{ margin: "0 0 8px" }}>{tx.archivePartsTitle}</h2>
              <p style={{ margin: "0 0 18px", color: COLORS.muted, lineHeight: 1.55 }}>{tx.archivePartsBody}</p>
              <div style={{ display: "flex", flexWrap: "wrap", gap: 10 }}>
                {watch.archive_parts.map(part => <a key={part.part_number} href={part.spotify_url} target="_blank" rel="noreferrer" style={{ ...secondaryButtonStyle, textDecoration: "none" }}>{tx.archivePart(part.part_number)} · {tx.archivePartTracks(part.track_count)}</a>)}
              </div>
            </section>

            <section style={{ marginTop: 34 }}>
              <div style={{ display: "flex", flexWrap: "wrap", gap: 15, alignItems: "center", justifyContent: "space-between", marginBottom: 15 }}>
                <h2 style={{ margin: 0 }}>{tx.tracksTitle}</h2>
                <label style={{ display: "flex", gap: 8, alignItems: "center", color: COLORS.muted, cursor: "pointer" }}><input type="checkbox" checked={removedOnly} onChange={event => { setOffset(0); setRemovedOnly(event.target.checked); }} />{tx.removedOnly}</label>
              </div>
              {page?.tracks.length === 0 && <div style={{ ...cardStyle, color: COLORS.muted, padding: 25 }}>{tx.noTracks}</div>}
              <div style={{ display: "grid", gap: 9 }}>
                {page?.tracks.map(track => {
                  const metadata = track.metadata;
                  return (
                    <article key={track.uri} style={{ ...cardStyle, padding: 12, display: "flex", flexWrap: "wrap", gap: 13, alignItems: "center" }}>
                      {metadata ? <a href={metadata.spotify_url} target="_blank" rel="noreferrer"><img src={metadata.image_url} alt={metadata.name} style={{ width: 58, height: 58, borderRadius: 7, objectFit: "cover", display: "block" }} /></a> : <div aria-hidden style={{ width: 58, height: 58, borderRadius: 7, background: COLORS.faint }} />}
                      <div style={{ flex: "1 1 220px", minWidth: 0 }}>
                        {metadata ? <><a href={metadata.spotify_url} target="_blank" rel="noreferrer" style={{ color: COLORS.text, fontWeight: 750, textDecoration: "none" }}>{metadata.name}</a><div style={{ color: COLORS.muted, fontSize: 13, marginTop: 4 }}><a href={metadata.artist_url} target="_blank" rel="noreferrer" style={{ color: "inherit" }}>{metadata.artists}</a></div></> : <strong>{tx.metadataUnavailable}</strong>}
                      </div>
                      <div style={{ flex: "0 1 210px", color: COLORS.muted, fontSize: 12, lineHeight: 1.6 }}>
                        <div>{tx.firstSeen}: {date(track.first_seen)}</div>
                        {track.removed_at && <div style={{ color: "#ffca80", fontWeight: 700 }}>{tx.removedOn}: {date(track.removed_at)}</div>}
                      </div>
                      <button type="button" onClick={() => handleDeleteTrack(track.uri)} disabled={deletingURI === track.uri} aria-label={tx.deleteTrackLabel(metadata?.name || tx.metadataUnavailable)} style={{ ...secondaryButtonStyle, padding: "8px 12px", color: "#ff9e9e" }}>{deletingURI === track.uri ? tx.deleting : tx.delete}</button>
                    </article>
                  );
                })}
              </div>
              {page && page.total > 0 && <div style={{ display: "flex", justifyContent: "center", alignItems: "center", flexWrap: "wrap", gap: 12, marginTop: 20 }}><button disabled={offset === 0} onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))} style={{ ...secondaryButtonStyle, opacity: offset === 0 ? 0.45 : 1 }}>{tx.previousPage}</button><span style={{ color: COLORS.muted, fontSize: 13 }}>{tx.pageStatus(pageFrom, pageTo, page.total)}</span><button disabled={offset + PAGE_SIZE >= page.total} onClick={() => setOffset(offset + PAGE_SIZE)} style={{ ...primaryButtonStyle, opacity: offset + PAGE_SIZE >= page.total ? 0.45 : 1 }}>{tx.nextPage}</button></div>}
            </section>
          </>
        )}
      </div>
    </Layout>
  );
}
