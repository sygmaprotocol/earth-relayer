package listener

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/events"
	"github.com/ChainSafe/chainbridge-core/communication"
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/tss"
	"github.com/ChainSafe/chainbridge-core/tss/keygen"
	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/rs/zerolog/log"
)

type EventListener interface {
	FetchDeposits(ctx context.Context, address common.Address, startBlock *big.Int, endBlock *big.Int) ([]*events.Deposit, error)
	FetchKeygenEvents(ctx context.Context, address common.Address, startBlock *big.Int, endBlock *big.Int) ([]ethTypes.Log, error)
	FetchRefreshEvents(ctx context.Context, address common.Address, startBlock *big.Int, endBlock *big.Int) ([]ethTypes.Log, error)
}

type DepositHandler interface {
	HandleDeposit(sourceID, destID uint8, nonce uint64, resourceID types.ResourceID, calldata, handlerResponse []byte) (*message.Message, error)
}

type DepositEventHandler struct {
	eventListener  EventListener
	depositHandler DepositHandler

	bridgeAddress common.Address
	domainID      uint8
}

func NewDepositEventHandler(eventListener EventListener, depositHandler DepositHandler, bridgeAddress common.Address, domainID uint8) *DepositEventHandler {
	return &DepositEventHandler{
		eventListener:  eventListener,
		depositHandler: depositHandler,
		bridgeAddress:  bridgeAddress,
		domainID:       domainID,
	}
}

func (eh *DepositEventHandler) HandleEvent(block *big.Int, msgChan chan *message.Message) error {
	deposits, err := eh.eventListener.FetchDeposits(context.Background(), eh.bridgeAddress, block, block)
	if err != nil {
		return fmt.Errorf("unable to fetch deposit events because of: %+v", err)
	}

	for _, d := range deposits {
		m, err := eh.depositHandler.HandleDeposit(eh.domainID, d.DestinationDomainID, d.DepositNonce, d.ResourceID, d.Data, d.HandlerResponse)
		if err != nil {
			log.Error().Str("block", block.String()).Uint8("domainID", eh.domainID).Msgf("%v", err)
			continue
		}

		log.Debug().Msgf("Resolved message %+v in block %s", m, block.String())
		msgChan <- m
	}

	return nil
}

type KeygenEventHandler struct {
	eventListener EventListener
	coordinator   *tss.Coordinator
	host          host.Host
	communication communication.Communication
	storer        keygen.SaveDataStorer
	bridgeAddress common.Address
	threshold     int
}

func NewKeygenEventHandler(eventListener EventListener, bridgeAddress common.Address) *KeygenEventHandler {
	return &KeygenEventHandler{
		eventListener: eventListener,
		bridgeAddress: bridgeAddress,
	}
}

func (eh *KeygenEventHandler) HandleEvent(block *big.Int, msgChan chan *message.Message) error {
	keygenEvents, err := eh.eventListener.FetchKeygenEvents(context.Background(), eh.bridgeAddress, block, block)
	if err != nil {
		return fmt.Errorf("unable to fetch keygen events because of: %+v", err)
	}
	if len(keygenEvents) == 0 {
		return nil
	}

	keygen := keygen.NewKeygen(block.String(), eh.threshold, eh.host, eh.communication, eh.storer)
	go eh.coordinator.Execute(context.Background(), keygen, make(chan interface{}, 1), make(chan error, 1))

	return nil
}

type RefreshEventHandler struct {
	eventListener EventListener
	bridgeAddress common.Address
}

func NewRefreshEventHandler(eventListener EventListener, bridgeAddress common.Address) *RefreshEventHandler {
	return &RefreshEventHandler{
		eventListener: eventListener,
		bridgeAddress: bridgeAddress,
	}
}

func (eh *RefreshEventHandler) HandleEvent(block *big.Int, msgChan chan *message.Message) error {
	refreshEvents, err := eh.eventListener.FetchRefreshEvents(context.Background(), eh.bridgeAddress, block, block)
	if err != nil {
		return fmt.Errorf("unable to fetch keygen events because of: %+v", err)
	}
	if len(refreshEvents) == 0 {
		return nil
	}

	return nil
}
