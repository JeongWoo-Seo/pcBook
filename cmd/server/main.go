package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/JeongWoo-Seo/pcBook/service"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	serverCertFile = "cert/server-cert.pem"
	serverKeyFile  = "cert/server-key.pem"
)

func main() {
	port := flag.Int("port", 0, "the server port")
	enableTls := flag.Bool("tls", false, "enable tls")
	serverType := flag.String("type", "grpc", "type of srver(grpc/rest)")
	endPoint := flag.String("endpoint", "", "grpc endpoint")
	flag.Parse()

	userStore := service.NewInMemoryUserStore()
	tokenManager := service.NewPasetoManager(service.TokenKey, service.TokenDuration)
	authServer := service.NewAuthServer(userStore, tokenManager)
	err := seedUser(userStore)
	if err != nil {
		log.Fatal("can not seed user")
	}

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("tmp")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("can not start server: ", err)
	}

	if *serverType == "grpc" {
		err = runGRPCServer(authServer, laptopServer, tokenManager, *enableTls, listener)
		if err != nil {
			log.Fatal("can not start grpc server: %w", err)
		}
	} else {
		err = runRESTServer(authServer, laptopServer, tokenManager, *enableTls, listener, *endPoint)
		if err != nil {
			log.Fatal("can not start REST server: %w", err)
		}
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	pemClientCA, err := os.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

func runGRPCServer(
	authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	tokenManager *service.PasetoManager,
	enableTLS bool,
	listener net.Listener,
) error {
	interceptor := service.NewAuthInterceptor(tokenManager, accessibleRole())
	serverOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	if enableTLS {
		tlsCerdentials, err := loadTLSCredentials()
		if err != nil {
			return fmt.Errorf("can not load tls credentials: %v", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(tlsCerdentials))
	}

	grpcServer := grpc.NewServer(serverOpts...)
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	reflection.Register(grpcServer)

	log.Printf("start GRPC server at %s, TLS = %t", listener.Addr().String(), enableTLS)
	return grpcServer.Serve(listener)
}

func runRESTServer(
	authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	tokenManager *service.PasetoManager,
	enableTLS bool,
	listener net.Listener,
	grpcEndpoint string,
) error {
	mux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts)
	if err != nil {
		return err
	}

	err = pb.RegisterLaptopServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts)
	if err != nil {
		return err
	}

	log.Printf("start REST server at %s, TLS = %t", listener.Addr().String(), enableTLS)

	if enableTLS {
		return http.ServeTLS(listener, mux, serverCertFile, serverKeyFile)
	}

	return http.Serve(listener, mux)
}

func seedUser(userStore service.UserStore) error {
	err := createUser(userStore, "admin", "secret", "admin")
	if err != nil {
		return err
	}

	return createUser(userStore, "user", "secret", "user")
}

func createUser(userStore service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}
	return userStore.Save(user)
}

func accessibleRole() map[string][]string {
	const laptopServicePath = "/pcbook.LaptopService/"

	return map[string][]string{
		laptopServicePath + "CreateLaptop": {"admin"},
		laptopServicePath + "UploadImage":  {"admin"},
		laptopServicePath + "RateLaptop":   {"admin", "user"},
	}
}
