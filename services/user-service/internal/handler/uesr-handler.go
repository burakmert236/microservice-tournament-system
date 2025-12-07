package handler

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/burakmert236/goodswipe-common/errors"
	proto "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	"github.com/burakmert236/goodswipe-user-service/internal/service"
)

type UserHandler struct {
	proto.UnimplementedUserServiceServer
	UserService service.UserService
}

func NewUserHandler(UserService service.UserService) *UserHandler {
	return &UserHandler{
		UserService: UserService,
	}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	if req.DisplayName == "" {
		return nil, status.Error(codes.InvalidArgument, "Display name is required")
	}

	user, err := h.UserService.CreateUser(ctx, req.DisplayName)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr.ToGRPCStatus()
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("internal server error: %s", err.Error()))
	}

	resp := &proto.CreateUserResponse{
		UserId: user.UserId,
	}

	return resp, nil
}

func (h *UserHandler) UpdateProgress(ctx context.Context, req *proto.UpdateProgressRequest) (*proto.MessageResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User id is required")
	}
	if req.ProgressAmount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Progress amount must be a positive number")
	}

	err := h.UserService.UpdateProgress(ctx, req.UserId, int(req.ProgressAmount))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr.ToGRPCStatus()
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("internal server error: %s", err.Error()))
	}

	message := &proto.MessageResponse{
		IsSuccess: true,
		Message:   "Update progress is succesful",
	}

	return message, nil
}
