package handler

import (
	"context"

	apperrors "github.com/burakmert236/goodswipe-common/errors"
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

func (h *TournamentHandler) EnterTournament(ctx context.Context, req *proto.EnterTournamentRequest) (*proto.EnterTournamentResponse, error) {
	if req.UserId == "" {
		return nil, apperrors.ToGRPCError(apperrors.New(apperrors.CodeInvalidInput, "user id is required"))
	}

	tournamentId, groupId, err := h.tournamentService.EnterTournament(ctx, req.UserId)
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	resp := &proto.EnterTournamentResponse{
		TournamentId: tournamentId,
		GroupId:      groupId,
	}

	return resp, nil
}

func (h *TournamentHandler) ClaimReward(ctx context.Context, req *proto.ClaimRewardRequest) (*proto.ClaimRewardResponse, error) {
	if req.UserId == "" || req.TournamentId == "" {
		return nil, apperrors.ToGRPCError(apperrors.New(apperrors.CodeInvalidInput, "user id and tournament id are required"))
	}

	tournamentId, reward, err := h.tournamentService.ClaimReward(ctx, req.UserId, req.TournamentId)
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	resp := &proto.ClaimRewardResponse{
		TournamentId: tournamentId,
		Reward:       int32(reward),
	}

	return resp, nil
}
