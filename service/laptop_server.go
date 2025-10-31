package service

import (
	"context"
	"errors"
	"log"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	Store LaptopStore
}

func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{Store: store}
}

func (s *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()

	log.Printf("receive create laptop request with id: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop id is not a valid id: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "can not create new laptop id: %v", err)
		}
		laptop.Id = id.String()
	}

	if ctx.Err() != nil {
		log.Printf("context error detected: %v", ctx.Err())
		return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded or canceled by client")
	}

	err := s.Store.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "cannot save laptop to the store: %v", err)
	}

	log.Printf("saved laptop with id: %s", laptop.Id)

	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}

	return res, nil
}

func (s *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream grpc.ServerStreamingServer[pb.SearchLaptopResponse]) error {
	filter := req.GetFilter()
	log.Printf("receive a search laptop with filter : %v", filter)

	err := s.Store.Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
		res := &pb.SearchLaptopResponse{Laptop: laptop}

		err := stream.Send(res)
		if err != nil {
			return err
		}

		log.Printf("send laptop with id: %s", laptop.GetId())
		return nil
	})

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}
