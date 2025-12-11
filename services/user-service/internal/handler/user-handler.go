package handler

import (
	"context"

	apperrors "github.com/burakmert236/goodswipe-common/errors"
	proto "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-user-service/internal/service"
)

type UserHandler struct {
	proto.UnimplementedUserServiceServer
	userService service.UserService
	logger      *logger.Logger
}

func NewUserHandler(UserService service.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: UserService,
		logger:      logger,
	}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	if req.DisplayName == "" {
		return nil, apperrors.ToGRPCError(apperrors.New(apperrors.CodeInvalidInput, "display name is required"))
	}

	user, err := h.userService.CreateUser(ctx, req.DisplayName)
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	resp := &proto.CreateUserResponse{
		UserId: user.UserId,
	}

	return resp, nil
}

func (h *UserHandler) UpdateProgress(ctx context.Context, req *proto.UpdateProgressRequest) (*proto.UpdateProgressResponse, error) {
	if req.UserId == "" {
		return nil, apperrors.ToGRPCError(apperrors.New(apperrors.CodeInvalidInput, "user id is required"))
	}
	if req.ProgressAmount <= 0 {
		return nil, apperrors.ToGRPCError(apperrors.New(apperrors.CodeInvalidInput, "Progress amount must be a positive number"))
	}

	user, err := h.userService.UpdateProgress(ctx, req.UserId, int(req.ProgressAmount))
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	message := &proto.UpdateProgressResponse{
		UserId: user.UserId,
		Level:  int32(user.Level),
		Coin:   int32(user.Coin),
	}

	return message, nil
}

func (h *UserHandler) GetById(ctx context.Context, req *proto.GetUserByIdRequest) (*proto.GetUserByIdResponse, error) {
	if req.UserId == "" {
		return nil, apperrors.ToGRPCError(apperrors.New(apperrors.CodeInvalidInput, "user id is required"))
	}

	user, err := h.userService.GetById(ctx, req.UserId)
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	message := &proto.GetUserByIdResponse{
		UserId:      user.UserId,
		DisplayName: user.DisplayName,
		Level:       int32(user.Level),
		Coin:        int32(user.Coin),
	}

	return message, nil
}

func (h *UserHandler) CollectTournamentReward(ctx context.Context, req *proto.CollectTournamentRewardRequest) (*proto.MessageResponse, error) {
	if req.UserId == "" || req.TournamentId == "" {
		return nil, apperrors.ToGRPCError(apperrors.New(apperrors.CodeInvalidInput, "user id is required"))
	}

	if req.Coin <= 0 {
		return nil, apperrors.ToGRPCError(apperrors.New(apperrors.CodeInvalidInput, "Reward must be a positive number"))
	}

	err := h.userService.CollectTournamentReward(ctx, req.UserId, req.TournamentId, int(req.Coin))
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	message := &proto.MessageResponse{
		IsSuccess: true,
		Message:   "Collecting coin for user is succesful",
	}

	return message, nil
}

func (h *UserHandler) ReserveCoins(ctx context.Context, req *proto.ReserveCoinsRequest) (*proto.MessageResponse, error) {
	err := h.userService.ReserveCoins(ctx, req.UserId, int(req.Amount), req.TournamentId)
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	return &proto.MessageResponse{
		IsSuccess: true,
		Message:   "coins reserved successfully",
	}, nil
}

func (h *UserHandler) ConfirmReservation(ctx context.Context, req *proto.ConfirmReservationRequest) (*proto.MessageResponse, error) {
	err := h.userService.ConfirmReservation(ctx, req.UserId, req.TournamentId)
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	return &proto.MessageResponse{
		IsSuccess: true,
		Message:   "reservation confirmed successfully",
	}, nil
}

func (h *UserHandler) RollbackReservation(ctx context.Context, req *proto.RollbackReservationRequest) (*proto.MessageResponse, error) {
	err := h.userService.RollbackReservation(ctx, req.UserId, req.TournamentId)
	if err != nil {
		return nil, apperrors.ToGRPCError(err)
	}

	return &proto.MessageResponse{
		IsSuccess: true,
		Message:   "reservation rollbacked successfully",
	}, nil
}
