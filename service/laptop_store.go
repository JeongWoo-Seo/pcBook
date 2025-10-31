package service

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"google.golang.org/protobuf/proto"
)

var ErrAlreadyExists = errors.New("record already exists")

type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

type inMemoryLaptopStore struct {
	mutax sync.RWMutex
	data  map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() *inMemoryLaptopStore {
	return &inMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

func (s *inMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	s.mutax.Lock()
	defer s.mutax.Unlock()

	if s.data[laptop.GetId()] != nil {
		return ErrAlreadyExists
	}

	other := proto.Clone(laptop).(*pb.Laptop)
	s.data[other.Id] = other

	return nil
}

func (s *inMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	s.mutax.Lock()
	defer s.mutax.Unlock()

	laptop := s.data[id]
	if laptop == nil {
		return nil, nil
	}

	other := proto.Clone(laptop).(*pb.Laptop)

	return other, nil
}

func (s *inMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
	s.mutax.Lock()
	defer s.mutax.Unlock()

	for _, laptop := range s.data {
		if ctx.Err() != nil {
			log.Printf("context error detected: %v", ctx.Err())
			return errors.New("deadline is exceeded or canceled by client")
		}

		if isQualified(filter, laptop) {
			other := proto.Clone(laptop).(*pb.Laptop)

			err := found(other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPrice() > filter.GetMaxPrice() {
		return false
	}

	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3
	case pb.Memory_KILOBYTE:
		return value << 13
	case pb.Memory_MEGABYTE:
		return value << 23
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}
