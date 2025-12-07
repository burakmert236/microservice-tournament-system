package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/burakmert236/goodswipe-tournament-service/internal/service"
)

type TournamentScheduler struct {
	tournamentService service.TournamentService
}

func NewTournamentScheduler(tournamentService service.TournamentService) *TournamentScheduler {
	return &TournamentScheduler{
		tournamentService: tournamentService,
	}
}

func (ts *TournamentScheduler) CreateTournament(ctx context.Context, now time.Time) error {
	log.Println("Creating daily tournament...")

	tournament, err := ts.tournamentService.CreateTournament(ctx, now)
	if err != nil {
		log.Printf("Failed to create tournament : %v", err)
	}

	log.Printf("Created tournament: (ID: %s)", tournament.TournamentId)

	return nil
}
