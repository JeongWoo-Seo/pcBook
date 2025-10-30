package service

import (
	"errors"
	"sync"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"google.golang.org/protobuf/proto"
)

var ErrAlreadyExists = errors.New("record already exists")

type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
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
