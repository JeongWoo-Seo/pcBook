package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/JeongWoo-Seo/pcBook/client"
	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/JeongWoo-Seo/pcBook/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	username        = "admin"
	password        = "secret"
	refreshDuration = 30 * time.Second
)

func main() {
	serverAddress := flag.String("address", "", "the server port")
	flag.Parse()
	log.Printf("server port : %s", *serverAddress)

	cc1, err := grpc.NewClient(*serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("can not create client: ", err)
	}

	authClient := client.NewAuthClinet(cc1, username, password)
	interceptor, err := client.NewAuthInterceptor(authClient, authmethods(), refreshDuration)
	if err != nil {
		log.Fatal("can not create auth interceptor: ", err)
	}
	cc2, err := grpc.NewClient(
		*serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal("can not create client: ", err)
	}
	laptopClient := client.NewLaptopClient(cc2)

	//testUploadImage(laptopClient)
	testRatingLaptop(laptopClient)
}

func authmethods() map[string]bool {
	const laptopServicePath = "/pcbook.LaptopService/"

	return map[string]bool{
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage":  true,
		laptopServicePath + "RateLaptop":   true,
	}
}

func testCreateLaptop(laptopClient *client.LaptopClient) {
	laptopClient.CreateLaptop(util.NewLaptop())
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
	for i := 0; i < 10; i++ {
		laptopClient.CreateLaptop(util.NewLaptop())
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
	laptopClient.SearchLaptop(filter)
}

func testUploadImage(laptopClient *client.LaptopClient) {
	laptop := util.NewLaptop()
	laptopClient.CreateLaptop(laptop)
	laptopClient.UploadImage(laptop.Id, "tmp/laptop.png")
}

func testRatingLaptop(laptopClient *client.LaptopClient) {
	n := 3
	laptopIDs := make([]string, 3)

	for i := 0; i < n; i++ {
		laptop := util.NewLaptop()
		laptopIDs[i] = laptop.Id
		laptopClient.CreateLaptop(laptop)
	}

	scores := make([]float64, n)
	for {
		fmt.Print("rate (y/n)? ")
		var answer string
		fmt.Scan(&answer)

		if answer != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = util.RandomLaptopScore()
		}

		err := laptopClient.RatingLaptop(laptopIDs, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}
