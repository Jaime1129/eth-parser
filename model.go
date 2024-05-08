package main

import "encoding/json"

// Transaction represents an Ethereum transaction
type Transaction struct {
	Hash string `json:"hash"`
	From string `json:"from"`
	To   string `json:"to"`
}

// Block represents an Ethereum block
type Block struct {
	Transactions []Transaction `json:"transactions"`
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
