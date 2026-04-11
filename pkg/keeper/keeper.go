package keeper

import "Aj-vrod/github-notifier/types"

type Keeper struct {
}

func NewKeeper() *Keeper {
	return &Keeper{}
}

func (k *Keeper) Store(info types.PRInfo) error {
	return nil
}
