package main

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	useradapter "todoe/domain/user/adapter"
	pb "todoe/onboarding"
)

type userGRPCServer struct {
	pb.UnimplementedOnboardingServiceServer
	repo *useradapter.MongoRepository
}

func (s *userGRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	id, err := bson.ObjectIDFromHex(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}
	result := s.repo.FindByID(ctx, id)
	if result.IsError() {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", result.Error())
	}
	u := result.MustGet()
	return &pb.GetUserResponse{
		UserId: u.ID.Hex(),
		Email:  u.Email,
		Name:   u.Name,
	}, nil
}
