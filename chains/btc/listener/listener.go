// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package listener

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/store"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/registry/parser"
	"github.com/rs/zerolog"
)

type EventHandler interface {
	HandleEvents(evts []btcjson.TxRawResult, msgChan chan []*message.Message) error
}

type ChainClient interface {
	LatestBlock() (*big.Int, error)
}

type ChainConnection interface {
	GetBestBlockHash()
	UpdateMetatdata() error
	GetHeaderLatest() (*types.Header, error)
	GetBlockHash(blockNumber uint64) (types.Hash, error)
	GetBlockEvents(hash types.Hash) ([]*parser.Event, error)
	GetFinalizedHead() (types.Hash, error)
	GetBlock(blockHash types.Hash) (*types.SignedBlock, error)
}

type BtcListener struct {
	conn rpcclient.Client

	eventHandlers      []EventHandler
	blockRetryInterval time.Duration

	log      zerolog.Logger
	domainID uint8
}

// NewBtcListener creates an BtcListener that listens to deposit events on chain
// and calls event handler when one occurs
func NewBtcListener(connection rpcclient.Client, eventHandlers []EventHandler /* config *substrate.SubstrateConfig, */, domainID uint8) *BtcListener {
	return &BtcListener{
		/* 		log:                log.With().Uint8("domainID", *config.GeneralChainConfig.Id).Logger(),
		 */conn:            connection,
		eventHandlers:      eventHandlers,
		blockRetryInterval: 10,
		domainID:           domainID,
	}
}

// ListenToEvents goes block by block of a network and executes event handlers that are
// configured for the listener.
func (l *BtcListener) ListenToEvents(ctx context.Context, startBlock *big.Int, domainID uint8, blockstore store.BlockStore, msgChan chan []*message.Message) {
	endBlock := big.NewInt(0)
	fmt.Println("listenameventeeeeeeee")
loop:
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Get the hash of the most recent block
			fmt.Println("evorkrecemememememe")
			bestBlockHash, err := l.conn.GetBestBlockHash()
			fmt.Println("bestblock:")
			fmt.Println(bestBlockHash)
			fmt.Println(err)
			if err != nil {
				l.log.Warn().Err(err).Msg("Unable to get latest block")
				time.Sleep(l.blockRetryInterval)
				return
			}
			fmt.Println("aaaaaaaaaaaaa")

			// Fetch the most recent block in verbose mode to get additional information including height
			block, err := l.conn.GetBlockVerboseTx(bestBlockHash)
			if err != nil {
				l.log.Warn().Err(err).Msg("Unable to get latest block")
				time.Sleep(l.blockRetryInterval)
				continue
			}
			head := big.NewInt(block.Height)
			if startBlock == nil {
				startBlock = head
			}

			fmt.Println("haeeaaddd")
			fmt.Println(startBlock.Cmp(head))
			// Sleep if the difference is less than needed block confirmations; (latest - current) < BlockDelay
			if startBlock.Cmp(head) == 1 {
				time.Sleep(l.blockRetryInterval)
				continue
			}
			fmt.Println("fetchamblokoveeehhahahsadsadsads")
			evts, err := l.fetchEvents(startBlock)
			if err != nil {
				l.log.Warn().Err(err).Msgf("Failed fetching events for block range %s-%s", startBlock, endBlock)
				time.Sleep(l.blockRetryInterval)
				continue
			}
			fmt.Println("evtssssssssssssssssssssssss")
			for _, handler := range l.eventHandlers {
				err := handler.HandleEvents(evts, msgChan)
				if err != nil {
					l.log.Warn().Err(err).Msgf("Unable to handle events")
					continue loop
				}
			}

			//Write to block store. Not a critical operation, no need to retry
			err = blockstore.StoreBlock(endBlock, l.domainID)
			if err != nil {
				l.log.Error().Str("block", endBlock.String()).Err(err).Msg("Failed to write latest block to blockstore")
			}

			startBlock.Add(startBlock, big.NewInt(1))
		}
	}
}

func (l *BtcListener) fetchEvents(startBlock *big.Int) ([]btcjson.TxRawResult, error) {
	l.log.Debug().Msgf("Fetching btc events for block %s", startBlock)

	blockHash, err := l.conn.GetBlockHash(startBlock.Int64())
	if err != nil {
		return nil, err
	}

	// Fetch block details in verbose mode
	block, err := l.conn.GetBlockVerboseTx(blockHash)
	if err != nil {
		return nil, err
	}
	return block.Tx, nil
}
