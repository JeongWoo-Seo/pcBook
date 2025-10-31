package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/JeongWoo-Seo/pcBook/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	serverAddress := flag.String("address", "", "the server port")
	flag.Parse()
	log.Printf("server port : %s", *serverAddress)

	con, err := grpc.NewClient(*serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("can not create client: ", err)
	}

	laptopClient := pb.NewLaptopServiceClient(con)

	for i := 0; i < 10; i++ {
		CreateLaptop(laptopClient)
	}

	filter := &pb.Filter{
		MaxPrice:    1500000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam: &pb.Memory{
			Value: 8,
			Unit:  pb.Memory_GIGABYTE,
		},
	}
	SearchLaptop(laptopClient, filter)
}

func CreateLaptop(laptopClient pb.LaptopServiceClient) {
	laptop := util.NewLaptop()
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("laptop already exists")
		} else {
			log.Print("can not create laptop: ", err)
		}
		return
	}
	log.Printf("created laptop with id: %s", res.Id)
}

func SearchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Printf("search filter: %v", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}

	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("fail to search laptop: ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("can not recieve response: ", err)
		}

		laptop := res.GetLaptop()
		log.Print("- found: ", laptop.GetId())
		log.Print("  + brand: ", laptop.GetBrand())
		log.Print("  + name: ", laptop.GetName())
		log.Print("  + cpu cores: ", laptop.GetCpu().GetNumberCores())
		log.Print("  + cpu min ghz: ", laptop.GetCpu().GetMinGhz())
		log.Print("  + ram: ", laptop.GetRam())
		log.Print("  + price: ", laptop.GetPrice())
	}

}
