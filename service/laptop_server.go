package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/JeongWoo-Seo/pcBook/redisutil"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxImageSize = 1 << 20
)

type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	LaptopStore LaptopStore
	ImageStore  ImageStore
	RatingStore RatingStore
	RDB         *redis.Client
}

func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore, rdb *redis.Client) *LaptopServer {
	return &LaptopServer{
		LaptopStore: laptopStore,
		ImageStore:  imageStore,
		RatingStore: ratingStore,
		RDB:         rdb,
	}
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

	if err := contextError(ctx); err != nil {
		return nil, err
	}

	err := s.LaptopStore.Save(laptop)
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

	err := s.LaptopStore.Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
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

func (s *LaptopServer) UploadImage(stream grpc.ClientStreamingServer[pb.UploadImageRequest, pb.UploadImageResponse]) error {
	req, err := stream.Recv()
	if err != nil {
		return logErr(status.Errorf(codes.Unknown, "can not recieve image info"))
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("recieve an image for laptop %s ", laptopID)

	laptop, err := s.LaptopStore.Find(laptopID)
	if err != nil {
		return logErr(status.Errorf(codes.Internal, "can not find laptop: %v", err))
	}
	if laptop == nil {
		return logErr(status.Errorf(codes.InvalidArgument, "laptop %s no exist", laptopID))
	}

	imageData := bytes.Buffer{}
	imagesie := 0

	for {
		if err := contextError(stream.Context()); err != nil {
			return err
		}
		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logErr(status.Errorf(codes.Unknown, "can not recieve chunk data: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		imagesie += size
		if imagesie > maxImageSize {
			return logErr(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imagesie, maxImageSize))
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			return logErr(status.Errorf(codes.Internal, "failed to write data: %v", err))
		}
	}

	imageId, err := s.ImageStore.Save(laptopID, imageType, imageData)
	if err != nil {
		return logErr(status.Errorf(codes.Internal, "can not save image data to store: %v", err))
	}

	res := &pb.UploadImageResponse{
		Id:   imageId,
		Size: uint32(imagesie),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logErr(status.Errorf(codes.Unknown, "failed to send res: %v", err))
	}

	log.Printf("saved image with id: %s", imageId)
	return nil
}

func (s *LaptopServer) RateLaptop(stream grpc.BidiStreamingServer[pb.RateLaptopRequest, pb.RateLaptopResponse]) error {
	for {
		if err := contextError(stream.Context()); err != nil {
			return err
		}
		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logErr(status.Errorf(codes.Unknown, "can not recieve data: %v", err))
		}

		laptopID := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("recieve rate laptop request: id = %s", laptopID)

		found, err := s.LaptopStore.Find(laptopID)
		if err != nil {
			return logErr(status.Errorf(codes.Internal, "can not find laptop: %v", err))
		}
		if found == nil {
			return logErr(status.Errorf(codes.NotFound, "laptop %s no exist", laptopID))
		}

		rating, err := s.RatingStore.Add(laptopID, score)
		if err != nil {
			return logErr(status.Errorf(codes.Internal, "can not add rating: %v", err))
		}

		res := &pb.RateLaptopResponse{
			LaptopId:     laptopID,
			RatedCount:   rating.Count,
			AverageScore: rating.sum / float64(rating.Count),
		}

		err = stream.Send(res)
		if err != nil {
			return logErr(status.Errorf(codes.Unknown, "can not send stream: %v", err))
		}
	}
	return nil
}

func (s *LaptopServer) SendLaptopInfo(stream grpc.ClientStreamingServer[pb.SendLaptopInfoRequest, pb.SendLaptopInfoResponse]) error {
	ctx := stream.Context()
	if err := contextError(ctx); err != nil {
		return err
	}
	var totalRecieved int

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return logErr(status.Errorf(codes.Unknown, "can not recieve laptop info data: %v", err))
		}

		log.Printf("Recieved laptop info to client : id = %v", req.GetLaptop().GetId())

		err = redisutil.PublishToRedis(ctx, s.RDB, req.GetLaptop())
		if err != nil {
			log.Printf("Warning: Failed to publish to Redis: %v", err)
		}

		totalRecieved++
	}

	res := &pb.SendLaptopInfoResponse{
		Msg: fmt.Sprintf("finished receiving %d laptop infos.", totalRecieved),
	}

	err := stream.SendAndClose(res)
	if err != nil {
		return logErr(status.Errorf(codes.Unknown, "failed to send res: %v", err))
	}

	return nil
}

func logErr(err error) error {
	if err != nil {
		log.Print(err)
	}

	return err
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logErr(status.Error(codes.Canceled, "canceled by client"))
	case context.DeadlineExceeded:
		return logErr(status.Error(codes.DeadlineExceeded, "req  DeadlineExceeded"))
	default:
		return nil
	}
}
