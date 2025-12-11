package handler

import (
	"context"

	apperrors "github.com/burakmert236/goodswipe-common/errors"
	proto "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-leaderboard-service/internal/service"
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
	leaderboard, err := h.leaderboardService.GetGlobalLeaderboard(ctx)
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	responseUsers := make([]*proto.UserInfo, len(leaderboard))
	for i, entry := range leaderboard {
		responseUsers[i] = &proto.UserInfo{
			UserId:      entry.UserId,
			DisplayName: entry.DisplayName,
			Score:       int64(entry.Score),
		}
	}

	return &proto.GetGlobalLeaderboardResponse{Users: responseUsers}, nil
}

func (h *LeaderboardHandler) GetTournamentLeaderboard(
	ctx context.Context,
	req *proto.GetTournamentLeaderboardRequest,
) (*proto.GetTournamentLeaderboardResponse, error) {
	if req.UserId == "" || req.TournamentId == "" {
		return nil, apperrors.ToGRPCError(apperrors.New(apperrors.CodeInvalidInput, "user id and tournament id is required"))
	}

	leaderboard, err := h.leaderboardService.GetTournamentLeaderboard(ctx, req.UserId, req.TournamentId)
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	responseUsers := make([]*proto.UserInfo, len(leaderboard))
	for i, entry := range leaderboard {
		responseUsers[i] = &proto.UserInfo{
			UserId:      entry.UserId,
			DisplayName: entry.DisplayName,
			Score:       int64(entry.Score),
		}
	}

	return &proto.GetTournamentLeaderboardResponse{Users: responseUsers}, nil
}

func (h *LeaderboardHandler) GetTournamentRank(
	ctx context.Context,
	req *proto.GetTournamentRankRequest,
) (*proto.GetTournamentRankResponse, error) {
	if req.UserId == "" || req.TournamentId == "" {
		return nil, apperrors.ToGRPCError(apperrors.New(apperrors.CodeInvalidInput, "user id and tournament id is required"))
	}

	rank, err := h.leaderboardService.GetTournamentRank(ctx, req.UserId, req.TournamentId)
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	return &proto.GetTournamentRankResponse{Rank: int32(rank)}, nil
}
