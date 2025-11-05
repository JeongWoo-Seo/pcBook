package service

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	tokenManager   *PasetoManager
	accessibleRole map[string][]string
}

func NewAuthInterceptor(tokenManager *PasetoManager, accessibleRole map[string][]string) *AuthInterceptor {
	return &AuthInterceptor{tokenManager: tokenManager, accessibleRole: accessibleRole}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		err := i.Authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		err := i.Authorize(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func (i *AuthInterceptor) Authorize(ctx context.Context, method string) error {
	accessibleRoles, ok := i.accessibleRole[method]
	if !ok {
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is empty")
	}

	value := md.Get(authorizationHeader)
	if len(value) == 0 {
		return status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := value[0]
	field := strings.Fields(authHeader)
	if len(field) < 2 {
		return status.Errorf(codes.Unauthenticated, "invalid auth header format")
	}

	authType := strings.ToLower(field[0])
	if authType != authorizationBearer {
		return status.Errorf(codes.Unauthenticated, "unsupported authorization type: %s", authType)
	}

	accessToken := field[1]
	payload, err := i.tokenManager.VerifyToken(accessToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid token")
	}

	for _, role := range accessibleRoles {
		if role == payload.Role {
			return nil
		}
	}
	return status.Errorf(codes.PermissionDenied, "no permission to access rpc")
}
