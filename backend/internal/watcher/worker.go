package watcher

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/shared/domain"
)

const workerConcurrency = 4

type dueWatchRepo interface {
	ClaimDue(ctx context.Context, interval time.Duration, limit int) ([]domain.Watch, error)
}

type pendingRepo interface {
	DeleteExpired(ctx context.Context) (int64, error)
}

type watchProcessor interface {
	ProcessWatch(ctx context.Context, watch domain.Watch) error
}

func NewWorker(watchRepo dueWatchRepo, pendingRepo pendingRepo, service watchProcessor) *Worker {
	return &Worker{watchRepo: watchRepo, pendingRepo: pendingRepo, service: service}
}

type Worker struct {
	watchRepo   dueWatchRepo
	pendingRepo pendingRepo
	service     watchProcessor
}

func (w *Worker) Start(ctx context.Context) error {
	ticker := time.NewTicker(domain.WatcherTickInterval)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return nil
		}
		w.tick(ctx)
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	watcherWatchesDue.Set(0)
	if _, err := w.pendingRepo.DeleteExpired(ctx); err != nil && ctx.Err() == nil {
		log.Error().Err(err).Msg("failed to delete expired pending selections")
	}
	if ctx.Err() != nil {
		return
	}

	watches, err := w.watchRepo.ClaimDue(ctx, domain.WatchPollInterval, domain.WatcherBatchSize)
	if err != nil {
		if ctx.Err() == nil {
			log.Error().Err(err).Msg("failed to claim due watches")
		}
		return
	}
	watcherWatchesDue.Set(float64(len(watches)))

	sem := make(chan struct{}, workerConcurrency)
	var wg sync.WaitGroup
	for _, watch := range watches {
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Wait()
			return
		}

		wg.Add(1)
		go func(watch domain.Watch) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := w.service.ProcessWatch(ctx, watch); err != nil && ctx.Err() == nil {
				log.Error().Err(err).Int("watch_id", watch.ID).Msg("failed to process watch")
			}
		}(watch)
	}
	wg.Wait()
}
