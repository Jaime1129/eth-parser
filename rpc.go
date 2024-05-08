package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type rpcClient struct {
	storage Storage
}

// Transaction represents an Ethereum transaction
type Transaction struct {
	Hash string `json:"hash"`
	From string `json:"from"`
	To   string `json:"to"`
}

// Block represents an Ethereum block
type Block struct {
	Transactions []Transaction `json:"transactions"`
	Number       string        `json:"number"`
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
	ID     int             `json:"id"`
}

// RPCError captures JSON-RPC error field
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *rpcClient) trackBlocks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			err := c.parseBlocks()
			if err != nil {
				fmt.Println("parse block err: " + err.Error())
			}
		case <-ctx.Done():
			fmt.Println("stop tracking blocks")
			return
		}
	}
}

func (c *rpcClient) parseBlocks() error {
	nodeURL := "https://eth.llamarpc.com" // Change this URL to the actual RPC endpoint

	// Create a new RPC request to get the latest block with full transaction details
	reqBody, err := json.Marshal(RPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_getBlockByNumber",
		Params:  []interface{}{"finalized", true},
		ID:      1,
	})
	if err != nil {
		return err
	}

	// Send the RPC request
	resp, err := http.Post(nodeURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return err
	}

	if rpcResp.Error != nil {
		fmt.Printf("Error from JSON-RPC: %s\n", rpcResp.Error.Message)
		return errors.New(rpcResp.Error.Message)
	}

	var block Block
	if err := json.Unmarshal(rpcResp.Result, &block); err != nil {
		return err
	}

	bjson, _ := json.Marshal(block)
	fmt.Println(string(bjson))

	num, err := HexToInt(block.Number)
	if err != nil {
		return err
	}
	c.storage.SetLatestBlock(num)
	fmt.Printf("parsed block %d\n", num)

	// fetch subsribers
	targetAddress := c.storage.GetSubsriberAddresses()
	if len(c.storage.GetSubsriberAddresses()) == 0 {
		fmt.Println("empty subscriber addresses")
		return nil
	}
	fmt.Printf("address map: %v\n", targetAddress)

	// Filter and save transactions related to subscribed addresses
	trxMap := make(map[string][]Transaction)
	for _, tx := range block.Transactions {
		if _, ok := targetAddress[tx.From]; ok {
			trxMap[tx.From] = append(trxMap[tx.From], tx)
			fmt.Printf("save trxs: address=%s, trxHash=%s\n", tx.From, tx.Hash)
		}
		if _, ok := targetAddress[tx.To]; ok {
			fmt.Printf("save trxs: address=%s, trxHash=%s\n", tx.To, tx.Hash)
			trxMap[tx.To] = append(trxMap[tx.To], tx)
		}
	}

	c.storage.SaveTrxs(trxMap)
	return nil
}

func HexToInt(hexStr string) (int64, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	// base 16 for hexadecimal, 64 bits
	decimalValue, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		fmt.Println("Error converting hex to decimal:", err)
		return 0, err
	}
	return decimalValue, nil
}
