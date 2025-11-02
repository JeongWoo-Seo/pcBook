package service

import "sync"

type RatingStore interface {
	Add(laptopID string, score float64) (*Rating, error)
}

type Rating struct {
	Count uint32
	sum   float64
}

type InmemoryRatingStore struct {
	mutax  sync.RWMutex
	rating map[string]*Rating
}

func NewInMemoryRatingStore() *InmemoryRatingStore {
	return &InmemoryRatingStore{
		rating: make(map[string]*Rating),
	}
}

func (store *InmemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	store.mutax.Lock()
	defer store.mutax.Unlock()

	rating := store.rating[laptopID]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			sum:   score,
		}
	} else {
		rating.Count++
		rating.sum += score
	}

	store.rating[laptopID] = rating
	return rating, nil
}
