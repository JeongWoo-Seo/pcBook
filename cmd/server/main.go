package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/JeongWoo-Seo/pcBook/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	tokenKey      = "cd56e76e8bf6a1c32eb26966c864e983"
	tokenDuration = 15 * time.Minute
)

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("start server port : %d", *port)

	userStore := service.NewInMemoryUserStore()
	tokenManager := service.NewPasetoManager(tokenKey, tokenDuration)
	authServer := service.NewAuthServer(userStore, tokenManager)
	err := seedUser(userStore)
	if err != nil {
		log.Fatal("can not seed user")
	}

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("tmp")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	interceptor := service.NewAuthInterceptor(tokenManager, accessibleRole())
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	)
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	reflection.Register(grpcServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("can not start server: ", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("can not start server: ", err)
	}
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
