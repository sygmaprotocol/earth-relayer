// The Licensed Work is (c) 2022 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package connection

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
)

type Connection struct {
	*rpcclient.Client
}

func NewBtcConnection(url string) (*Connection, error) {

	// Connect to a Bitcoin node using RPC
	connConfig := &rpcclient.ConnConfig{
		HTTPPostMode: true,
		Host:         url,
		DisableTLS:   false,
	}

	client, err := rpcclient.New(connConfig, nil)
	if err != nil {
		fmt.Println("uerorusam")
		fmt.Println(err)
		log.Fatal(err)
	}

	return &Connection{
		Client: client,
	}, nil
}
