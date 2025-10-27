package util

import (
	"github.com/JeongWoo-Seo/pcBook/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewKeyboard() *pb.Keyboard {
	return &pb.Keyboard{
		Layout:  RandomKeyboardLayout(),
		Backlit: RandomBool(),
	}
}

func NewCPU() *pb.CPU {
	brand := RandomCPUBrand()
	numberCores := RandomInt(2, 8)
	minGhz := RandomFloat64(2.0, 3.5)

	return &pb.CPU{
		Brand:         brand,
		Name:          RandomCPUName(brand),
		NumberCores:   numberCores,
		NumberThreads: RandomInt(numberCores, 12),
		MinGhz:        minGhz,
		MaxGhz:        RandomFloat64(minGhz, 5.0),
	}
}

func NewGPU() *pb.GPU {
	brand := RandomGPUBrand()
	minGhz := RandomFloat64(1.0, 1.5)

	return &pb.GPU{
		Brand:  brand,
		Name:   RandomGPUName(brand),
		MinGhz: minGhz,
		MaxGhz: RandomFloat64(minGhz, 2.0),
		Memory: &pb.Memory{
			Value: uint64(RandomInt(2, 6)),
			Unit:  pb.Memory_GIGABYTE,
		},
	}
}

func NewRam() *pb.Memory {
	return &pb.Memory{
		Value: uint64(RandomInt(4, 64)),
		Unit:  pb.Memory_GIGABYTE,
	}
}

func NewSSD() *pb.Storage {
	return &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: &pb.Memory{
			Value: uint64(RandomInt(128, 1024)),
			Unit:  pb.Memory_GIGABYTE,
		},
	}
}

func NewHDD() *pb.Storage {
	return &pb.Storage{
		Driver: pb.Storage_HDD,
		Memory: &pb.Memory{
			Value: uint64(RandomInt(128, 1024)),
			Unit:  pb.Memory_TERABYTE,
		},
	}
}

func NewScreen() *pb.Screen {
	return &pb.Screen{
		SizeInch:   RandomFloat32(13, 17),
		Resolution: RandomResolution(),
		Panel:      RandomScreenPanel(),
		Multitouch: RandomBool(),
	}
}

func NewLaptop() *pb.Laptop {
	brand := RandomLaptopBrand()

	return &pb.Laptop{
		Id:       RandomID(),
		Brand:    brand,
		Name:     RandomLaptopName(brand),
		Cpu:      NewCPU(),
		Ram:      NewRam(),
		Gpus:     []*pb.GPU{NewGPU()},
		Storages: []*pb.Storage{NewSSD(), NewHDD()},
		Screen:   NewScreen(),
		Keyboard: NewKeyboard(),
		Weight: &pb.Laptop_WeightKg{
			WeightKg: RandomFloat64(1.0, 3.0),
		},
		Price:       RandomInt(1000000, 2000000),
		ReleaseYear: RandomInt(2022, 2025),
		UpdatedAt:   timestamppb.Now(),
	}
}
