package scheduler

import (
	"context"
	"log"
	"time"
)

type Scheduler struct {
	tournamentScheduler *TournamentScheduler
	stopChan            chan struct{}
}

func NewScheduler(
	tournamentScheduler *TournamentScheduler,
) *Scheduler {
	return &Scheduler{
		tournamentScheduler: tournamentScheduler,
		stopChan:            make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	ctx := context.Background()
	s.tournamentScheduler.CreateCurrentTournamentIfNotExists(ctx)

	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	durationUntilMidnight := nextMidnight.Sub(now)

	log.Printf("Next tournament creation scheduled at: %s (in %v)",
		nextMidnight.Format(time.RFC3339), durationUntilMidnight)

	timer := time.NewTimer(durationUntilMidnight)

	for {
		select {
		case <-timer.C:
			now := time.Now()
			ctx := context.Background()
			log.Println("Creating daily tournament at 00:00 UTC...")

			if err := s.tournamentScheduler.CreateTournament(ctx, now); err != nil {
				log.Printf("ERROR: Failed to create daily tournament: %v", err)
			} else {
				log.Println("Successfully created daily tournament")
			}

			timer.Reset(24 * time.Hour)

		case <-s.stopChan:
			timer.Stop()
			log.Println("Tournament creation scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) Stop() error {
	close(s.stopChan)
	return nil
}
