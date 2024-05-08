package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Parser interface {

	// last parsed block
	GetCurrentBlock() int

	// add address to observer
	Subscribe(address string) bool

	// list of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

type parser struct {
	storage Storage
}

func (p *parser) GetCurrentBlock() int {
	return int(p.storage.GetLatestBlock())
}

func (p *parser) Subscribe(address string) bool {
	if !isValidEthereumAddress(address) {
		fmt.Println("invalid ethereum address")
		return false
	}
	p.storage.AddSubcriber(address)
	return true
}

func (p *parser) GetTransactions(address string) []Transaction {
	if !isValidEthereumAddress(address) {
		return []Transaction{}
	}
	return p.storage.GetTrxsByAddress(address)
}

// isValidEthereumAddress checks if the provided string is a valid Ethereum address.
func isValidEthereumAddress(address string) bool {
	// Check if the address starts with '0x'
	if !strings.HasPrefix(address, "0x") {
		return false
	}

	// Remove the '0x' prefix
	address = address[2:]

	// Check if the address has exactly 40 hex characters
	if len(address) != 40 {
		return false
	}

	// Check if the address is valid hexadecimal
	if _, err := hex.DecodeString(address); err != nil {
		return false
	}
	return true
}

func setupServer(storage Storage) *http.Server {
	parser := &parser{storage: storage}
	sm := http.NewServeMux()
	sm.HandleFunc("/current-block", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
			return
		}
		num := parser.GetCurrentBlock()
		resp, err := json.Marshal(map[string]interface{}{"block_number": num})
		if err != nil {
			http.Error(w, "Marshal err", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(resp))
	})

	sm.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}
		// Read the body of the request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// Unmarshal the JSON data
		var data map[string]string
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "Error parsing JSON data", http.StatusBadRequest)
			return
		}

		res := parser.Subscribe(data["address"])
		resp, err := json.Marshal(map[string]bool{"success": res})
		if err != nil {
			http.Error(w, "Marshal err", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(resp))
	})
	sm.HandleFunc("/get-transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
			return
		}
		addr := r.URL.Query().Get("address")
		trxs := parser.GetTransactions(addr)
		resp, err := json.Marshal(map[string][]Transaction{"trxs": trxs})
		if err != nil {
			http.Error(w, "Marshal err", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(resp))
	})
	return &http.Server{
		Addr:    ":8080",
		Handler: sm,
	}
}
