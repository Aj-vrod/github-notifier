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

func (s *Storage) Subscribe(prInfo *types.PRInfo, prState types.PRState) {
	log.Printf("Subscribing new PR: %s/%s #%d", prInfo.Owner, prInfo.Repo, prInfo.Number)
	s.registry[prInfo.URL] = prState

}

func (s *Storage) Unsubscribe(prInfo *types.PRInfo) {
	log.Printf("Unsubcribing PR: %s/%s #%d", prInfo.Owner, prInfo.Repo, prInfo.Number)
	delete(s.registry, prInfo.URL)
}

func (s *Storage) GetAllSubscriptions() types.Registry {
	return s.registry
}
