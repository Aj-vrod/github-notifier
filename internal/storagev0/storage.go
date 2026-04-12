package storagev0

import (
	"Aj-vrod/github-notifier/types"
	"log"
)

type Storage struct {
	registry types.Registry
}

func NewStorage() *Storage {
	return &Storage{
		registry: make(types.Registry),
	}
}

func (s *Storage) AddSubscription(prInfo *types.PRInfo, prState types.PRState) {
	log.Printf("Subscribing new PR URL: %s", prInfo.URL)
	s.registry[prInfo.URL] = prState

}

func (s *Storage) GetAllSubscriptions() types.Registry {
	return s.registry
}
