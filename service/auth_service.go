package service

import (
	"context"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	UserStore    UserStore
	TokenManager *PasetoManager
}

func NewAuthServer(userStore UserStore, tokenManager *PasetoManager) *AuthServer {
	return &AuthServer{
		UserStore:    userStore,
		TokenManager: tokenManager,
	}
}

func (server *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := server.UserStore.Find(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can no find user: %v", err)
	}

	if user == nil || !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.InvalidArgument, "incorrect user info")
	}

	token, err := server.TokenManager.CreateToken(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create token: %v", err)
	}

	res := &pb.LoginResponse{AccessToken: token}
	return res, nil
}
