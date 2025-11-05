package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct {
	authClient  *AuthClient
	authMethods map[string]bool
	accessToken string
}

func NewAuthInterceptor(
	authClinet *AuthClient,
	authMethods map[string]bool,
	refreshDuration time.Duration,
) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		authClient:  authClinet,
		authMethods: authMethods,
	}

	err := interceptor.scheduleRefreshToken(refreshDuration)
	if err != nil {
		return nil, err
	}
	return interceptor, nil
}

func (i *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {
		if i.authMethods[method] {
			return invoker(i.attachToken(ctx), method, req, reply, cc, opts...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (i *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		if i.authMethods[method] {
			return streamer(i.attachToken(ctx), desc, cc, method, opts...)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (i *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	bearerToken := fmt.Sprintf("Bearer %s", i.accessToken)
	return metadata.AppendToOutgoingContext(ctx, "authorization", bearerToken)
}

func (i *AuthInterceptor) scheduleRefreshToken(refreshDuration time.Duration) error {
	err := i.refreshToken()
	if err != nil {
		return err
	}

	go func() {
		wait := refreshDuration
		for {
			time.Sleep(wait)
			err := i.refreshToken()
			if err != nil {
				wait = time.Second
			} else {
				wait = refreshDuration
			}
		}
	}()

	return nil
}

func (i *AuthInterceptor) refreshToken() error {
	accessToken, err := i.authClient.Login()
	if err != nil {
		return err
	}
	i.accessToken = accessToken
	log.Printf("token refreshed: %v", accessToken)
	return nil
}
