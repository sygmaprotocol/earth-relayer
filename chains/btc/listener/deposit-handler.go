// The Licensed Work is (c) 2022 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package listener

import (
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	core_types "github.com/ChainSafe/chainbridge-core/types"
)

type DepositHandlerFunc func(sourceID uint8, calldata []byte) (*message.Message, error)

const (
	PermissionlessGenericTransfer message.TransferType = "PermissionlessGenericTransfer"
)

type BtcDepositHandler struct {
	depositHandler DepositHandlerFunc
}

const (
	FungibleTransfer = iota
)
const FUNC_SIGNATURE = ""
const ADAPTER_CONTRACT_ADDRESS = ""
const MAX_FEE = ""
const DESTINATION_ID = 1

var NONCE = 1

var RESOURCE_ID = core_types.ResourceID{0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04}

// NewBtcDepositHandler creates an instance of BtcDepositHandler that contains
// handler functions for processing deposit events
func NewBtcDepositHandler() *BtcDepositHandler {
	return &BtcDepositHandler{}
}

func (e *BtcDepositHandler) HandleDeposit(sourceID uint8, depositorAddress string, calldata []byte) (*message.Message, error) {
	evmAddress := calldata[:20]

	payload := []interface{}{
		FUNC_SIGNATURE,
		ADAPTER_CONTRACT_ADDRESS,
		MAX_FEE,
		depositorAddress,
		evmAddress,
	}

	metadata := message.Metadata{
		Data: make(map[string]interface{}),
	}

	return message.NewMessage(sourceID, DESTINATION_ID, uint64(NONCE), RESOURCE_ID, PermissionlessGenericTransfer, payload, metadata), nil
}
