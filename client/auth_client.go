package client

import (
	"context"
	"time"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"google.golang.org/grpc"
)

type AuthClient struct {
	service  pb.AuthServiceClient
	username string
	password string
}

func NewAuthClinet(cc *grpc.ClientConn, username, password string) *AuthClient {
	service := pb.NewAuthServiceClient(cc)
	return &AuthClient{service: service, username: username, password: password}
}

func (client *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.LoginRequest{
		Username: client.username,
		Password: client.password,
	}

	res, err := client.service.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetAccessToken(), nil
}
