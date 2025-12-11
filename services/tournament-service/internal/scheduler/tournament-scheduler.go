package scheduler

import (
	"context"
	"log"
	"time"

	apperrors "github.com/burakmert236/goodswipe-common/errors"
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

func (ts *TournamentScheduler) CreateTournament(ctx context.Context, startsAt time.Time) *apperrors.AppError {
	log.Println("Creating daily tournament...")

	tournament, err := ts.tournamentService.CreateTournament(ctx, startsAt)
	if err != nil {
		log.Printf("Failed to create tournament : %v", err)
	}

	log.Printf("Created tournament: (ID: %s)", tournament.TournamentId)

	return nil
}

func (ts *TournamentScheduler) CreateCurrentTournamentIfNotExists(ctx context.Context) *apperrors.AppError {
	log.Println("Creating current tournament if not exists...")

	tournament, err := ts.tournamentService.CreateCurrentTournament(ctx)
	if err != nil {
		log.Printf("Failed to create tournament : %v", err)
	}

	log.Printf("Current tournament: (ID: %s)", tournament.TournamentId)

	return nil
}
