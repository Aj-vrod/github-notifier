package storagev0

import (
	"Aj-vrod/github-notifier/types"
	"fmt"
)

type Storage struct {
	registry types.Registry
}

func NewStorage() *Storage {
	return &Storage{
		registry: make(types.Registry),
	}
}

func (s *Storage) AddSubscription(prURL string, prState types.PRState) {
	s.registry[prURL] = prState

	// for testing
	fmt.Println(">>>", s.registry)
}
