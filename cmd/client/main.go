package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/JeongWoo-Seo/pcBook/client"
	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/JeongWoo-Seo/pcBook/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	username        = "admin"
	password        = "secret"
	refreshDuration = 30 * time.Second
)

func main() {
	serverAddress := flag.String("address", "", "the server port")
	enableTls := flag.Bool("tls", false, "enable tls")
	flag.Parse()
	log.Printf("server port : %s", *serverAddress)

	transferOption := grpc.WithTransportCredentials(insecure.NewCredentials())

	if *enableTls {
		tlscredentials, err := loadTLSCredentials()
		if err != nil {
			log.Fatal("can not laod tls credentials: ", err)
		}
		transferOption = grpc.WithTransportCredentials(tlscredentials)
	}

	cc1, err := grpc.NewClient(*serverAddress, transferOption)
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
		transferOption,
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal("can not create client: ", err)
	}
	laptopClient := client.NewLaptopClient(cc2)

	//testRatingLaptop(laptopClient)
	err = GetPcBookInfo(laptopClient)
	if err != nil {
		log.Fatal("failed to get PC info:", err)
	}
}

func authmethods() map[string]bool {
	const laptopServicePath = "/pcbook.LaptopService/"

	return map[string]bool{
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage":  true,
		laptopServicePath + "RateLaptop":   true,
	}
}

func GetPcBookInfo(laptopClient *client.LaptopClient) error {
	sendQueue := make(chan *pb.LaptopInfo, 100)

	go client.StartSenderWorker(laptopClient, sendQueue)

	defer close(sendQueue)
	defer time.Sleep(1 * time.Second)

	laptopId, err := laptopClient.GetMacSerialID()
	if err != nil {
		return fmt.Errorf("can not get laptop id")
	}

	for {
		start := time.Now()

		var wg sync.WaitGroup
		wg.Add(4)

		result := &pb.LaptopInfo{}

		errs := make(chan error, 4)

		// 정보 수집 - GetCPUInfo 와 GetNetInfo에서 1초 동안 데이터를 수집 for loop는 1초단위로 동작
		go func() {
			defer wg.Done()
			info, err := laptopClient.GetBatteryInfo()
			if err != nil {
				errs <- fmt.Errorf("battery info failed: %w", err)
				return
			}
			result.Battery = info
		}()

		go func() {
			defer wg.Done()
			info, err := laptopClient.GetCPUInfo()
			if err != nil {
				errs <- fmt.Errorf("cpu info failed: %w", err)
				return
			}
			result.Cpu = info
		}()

		go func() {
			defer wg.Done()
			ram, storage, err := laptopClient.GetMemoryInfo()
			if err != nil {
				errs <- fmt.Errorf("ram/disk info failed: %w", err)
				return
			}
			result.Ram = ram
			result.Storages = storage
		}()

		go func() {
			defer wg.Done()
			info, err := laptopClient.GetNetInfo()
			if err != nil {
				errs <- fmt.Errorf("net info failed: %w", err)
				return
			}
			result.Network = info
		}()

		wg.Wait() // 1초 단위로 정보 수집
		close(errs)

		// error
		for err := range errs {
			log.Printf("System info collection error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		result.Id = laptopId
		result.CreateAt = timestamppb.Now()

		elapsed := time.Since(start)

		fmt.Printf("\n===== 수집 완료 (%s) =====\n", elapsed)
		fmt.Printf("laptop id: %s\n", result.GetId())
		fmt.Printf("Battery: %d\n", result.GetBattery())
		fmt.Printf("CPU: %f\n", result.GetCpu())
		fmt.Printf("RAM: %s\n", result.GetRam())
		fmt.Printf("storage: %s\n", result.GetStorages())
		fmt.Printf("Net: rx: %d,tx: %d\n", result.GetNetwork().GetRx(), result.GetNetwork().GetTx())

		sendQueue <- result
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

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	ClientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{ClientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}
