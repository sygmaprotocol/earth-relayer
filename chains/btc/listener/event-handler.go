// The Licensed Work is (c) 2022 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package listener

import (
	"encoding/hex"
	"fmt"

	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DepositHandler interface {
	HandleDeposit(sourceID uint8, bridgeAddress string, calldata []byte) (*message.Message, error)
}

type FungibleTransferEventHandler struct {
	depositHandler DepositHandler
	resourceID     types.ResourceID
	domainID       uint8
	log            zerolog.Logger
}

func NewFungibleTransferEventHandler() *FungibleTransferEventHandler {
	return &FungibleTransferEventHandler{}
}

const BRIDGE_ADDRESS = ""

func (eh *FungibleTransferEventHandler) HandleEvents(evts []btcjson.TxRawResult, msgChan chan []*message.Message) error {
	domainDeposits := make(map[uint8][]*message.Message)

	for _, evt := range evts {
		for _, vout := range evt.Vout {
			fmt.Println(vout.ScriptPubKey.Type)
			fmt.Println(vout.ScriptPubKey.Hex)

			fmt.Println("adresa")
			fmt.Println(vout.ScriptPubKey.Address)

			if vout.ScriptPubKey.Address == BRIDGE_ADDRESS {
				func(evt btcjson.Vout) {
					defer func() {
						if r := recover(); r != nil {
							log.Error().Msgf("panic occured while handling deposit %+v", evt)
						}
					}()

					if vout.ScriptPubKey.Type == "op_return" {
						fmt.Println("  This is an OP_RETURN output")
						opReturnData := vout.ScriptPubKey.Hex

						// Decode the hexadecimal data if needed
						decodedData, err := hex.DecodeString(opReturnData)
						if err != nil {
							fmt.Println("Error decoding OP_RETURN data:", err)
						} else {
							fmt.Println("  Decoded Data:", string(decodedData))
						}
					}
					btcAddress := vout.ScriptPubKey.Address
					amount := vout.Value
					fmt.Println(btcAddress)
					fmt.Println(amount)
					callData, err := hex.DecodeString(vout.ScriptPubKey.Hex)
					if err != nil {
						eh.log.Error().Err(err).Msgf("Failed handling deposit %+v", vout.ScriptPubKey.Hex)
						return
					}
					m, err := eh.depositHandler.HandleDeposit(eh.domainID, BRIDGE_ADDRESS, callData)
					if err != nil {
						log.Error().Err(err).Msgf("%v", err)
						return
					}

					eh.log.Info().Msgf("Resolved deposit message %+v", vout.ScriptPubKey.Hex)

					domainDeposits[m.Destination] = append(domainDeposits[m.Destination], m)
				}(vout)
			}

		}

	}

	for _, deposits := range domainDeposits {
		go func(d []*message.Message) {
			msgChan <- d
		}(deposits)
	}
	return nil
}
