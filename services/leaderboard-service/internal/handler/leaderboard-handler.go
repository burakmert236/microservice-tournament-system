package handler

import (
	"context"
	"fmt"

	"github.com/burakmert236/goodswipe-common/errors"
	proto "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-leaderboard-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LeaderboardHandler struct {
	proto.UnimplementedLeaderboardServiceServer
	leaderboardService service.LeaderboardService
	logger             *logger.Logger
}

func NewLeaderboardHandler(leaderboardService service.LeaderboardService, logger *logger.Logger) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
		logger:             logger,
	}
}

func (h *LeaderboardHandler) GetGlobalLeaderboard(
	ctx context.Context,
	req *proto.GetGlobalLeaderboardRequest,
) (*proto.GetGlobalLeaderboardResponse, error) {
	err := h.leaderboardService.GetGlobalLeaderboard(ctx)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr.ToGRPCStatus()
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("internal server error: %v", err))
	}

	resp := &proto.GetGlobalLeaderboardResponse{
		// UserId: user.UserId,
	}

	return resp, nil
}

func (h *LeaderboardHandler) GetTournamentLeaderboard(
	ctx context.Context,
	req *proto.GetTournamentLeaderboardRequest,
) (*proto.GetTournamentLeaderboardResponse, error) {
	if req.UserId == "" || req.TournamentId == "" {
		return nil, status.Error(codes.InvalidArgument, "User id and tournament id is required")
	}

	err := h.leaderboardService.GetTournamentLeaderboard(ctx, req.UserId, req.TournamentId)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr.ToGRPCStatus()
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("internal server error: %v", err))
	}

	resp := &proto.GetTournamentLeaderboardResponse{
		// UserId: user.UserId,
	}

	return resp, nil
}

func (h *LeaderboardHandler) GetTournamentRank(
	ctx context.Context,
	req *proto.GetTournamentRankRequest,
) (*proto.GetTournamentRankResponse, error) {
	if req.UserId == "" || req.TournamentId == "" {
		return nil, status.Error(codes.InvalidArgument, "User id and tournament id is required")
	}

	err := h.leaderboardService.GetTournamentRank(ctx, req.UserId, req.TournamentId)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr.ToGRPCStatus()
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("internal server error: %v", err))
	}

	resp := &proto.GetTournamentRankResponse{
		// UserId: user.UserId,
	}

	return resp, nil
}
