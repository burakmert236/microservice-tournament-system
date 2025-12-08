package handler

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/burakmert236/goodswipe-common/errors"
	proto "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-tournament-service/internal/service"
)

type TournamentHandler struct {
	proto.UnimplementedTournamentServiceServer
	tournamentService service.TournamentService
	logger            *logger.Logger
}

func NewTournamentHandler(TournamentService service.TournamentService, logger *logger.Logger) *TournamentHandler {
	return &TournamentHandler{
		tournamentService: TournamentService,
		logger:            logger,
	}
}

func (h *TournamentHandler) EnterTournament(ctx context.Context, req *proto.EnterTournamentRequest) (*proto.MessageResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User id is required")
	}

	err := h.tournamentService.EnterTournament(ctx, req.UserId, req.UserId)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr.ToGRPCStatus()
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("internal server error: %v", err))
	}

	resp := &proto.MessageResponse{
		IsSuccess: true,
		Message:   "User registered to tournament succesfully",
	}

	return resp, nil
}

func (h *TournamentHandler) UpdateProgress(ctx context.Context, req *proto.ClaimRewardRequest) (*proto.MessageResponse, error) {
	return nil, nil
}
