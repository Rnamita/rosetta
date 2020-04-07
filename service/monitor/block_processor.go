package monitor

import (
	"context"

	"github.com/celo-org/rosetta/celo/client"
	"github.com/celo-org/rosetta/celo/wrapper"
	"github.com/celo-org/rosetta/db"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/prometheus/common/log"
)

func BlockProcessor(ctx context.Context, headers <-chan *types.Header, changes chan<- *db.BlockChangeSet, cc *client.CeloClient, db_ db.RosettaDBReader) error {

	registry, err := wrapper.NewRegistry(cc)
	if err != nil {
		return err
	}

	for h := range headers {
		bcs := db.BlockChangeSet{
			BlockNumber: h.Number,
		}

		blockNumber := h.Number.Uint64()
		iter, err := registry.Contract().FilterRegistryUpdated(&bind.FilterOpts{
			End:     &blockNumber,
			Start:   blockNumber,
			Context: ctx,
		}, nil)

		registryChanges := make([]db.RegistryChange, 0)
		for iter.Next() {
			registryChanges = append(registryChanges, db.RegistryChange{
				TxIndex:    iter.Event.Raw.TxIndex,
				Contract:   iter.Event.Identifier,
				NewAddress: iter.Event.Addr,
			})
			log.Info("Core Contract Address Changed", "name", iter.Event.Identifier, "newAddress", iter.Event.Addr, "txIndex", iter.Event.Raw.TxIndex)
		}
		if err != nil {
			return err
		}

		bcs.RegistryChanges = registryChanges

		select {
		case <-ctx.Done():
			return ctx.Err()
		case changes <- &bcs:
		}
	}

	return nil
}
