package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RPC config
const (
	rpcURL    = "http://127.0.0.1:18443"
	rpcWallet = "http://127.0.0.1:18443/wallet/juma"
	rpcUser   = "bootcamp"
	rpcPass   = "bootcamp123"
)

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
	ID     int             `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func callRPC(method string, params []interface{}, result interface{}) error {
	reqBody, _ := json.Marshal(RPCRequest{
		JSONRPC: "1.0",
		Method:  method,
		Params:  params,
		ID:      1,
	})

	req, _ := http.NewRequest("POST", rpcURL, bytes.NewBuffer(reqBody))
	req.SetBasicAuth(rpcUser, rpcPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return fmt.Errorf("json error: %w", err)
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return json.Unmarshal(rpcResp.Result, result)
}

func callWalletRPC(method string, params []interface{}, result interface{}) error {
	reqBody, _ := json.Marshal(RPCRequest{
		JSONRPC: "1.0",
		Method:  method,
		Params:  params,
		ID:      1,
	})

	req, _ := http.NewRequest("POST", rpcWallet, bytes.NewBuffer(reqBody))
	req.SetBasicAuth(rpcUser, rpcPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return fmt.Errorf("json error: %w", err)
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return json.Unmarshal(rpcResp.Result, result)
}

// Challenge 1: Blockchain Info
type BlockchainInfo struct {
	Chain      string  `json:"chain"`
	Blocks     int     `json:"blocks"`
	Difficulty float64 `json:"difficulty"`
}

func getBlockchainInfo() {
	var info BlockchainInfo
	if err := callRPC("getblockchaininfo", []interface{}{}, &info); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("\n--- Blockchain Info ---")
	fmt.Printf("Chain:      %s\n", info.Chain)
	fmt.Printf("Blocks:     %d\n", info.Blocks)
	fmt.Printf("Difficulty: %f\n", info.Difficulty)
}

// Challenge 2: Wallet Balance
func getWalletBalance() {
	var balance float64
	if err := callWalletRPC("getbalance", []interface{}{"*", 1}, &balance); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("\n--- Wallet Balance ---")
	fmt.Printf("Balance: %.8f BTC\n", balance)
}

// Challenge 3: List Recent Transactions
type Transaction struct {
	TXID     string  `json:"txid"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Time     int64   `json:"time"`
}

func listTransactions() {
	var txs []Transaction
	if err := callWalletRPC("listtransactions", []interface{}{"*", 5}, &txs); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("\n--- Recent Transactions ---")
	for _, tx := range txs {
		fmt.Printf("TXID:     %s\n", tx.TXID)
		fmt.Printf("Category: %s\n", tx.Category)
		fmt.Printf("Amount:   %.8f BTC\n", tx.Amount)
		fmt.Println("---")
	}
}

// Challenge 4: Decode a Transaction
type Vin struct {
	Coinbase string `json:"coinbase"`
	TXID     string `json:"txid"`
	Vout     int    `json:"vout"`
}

type Vout struct {
	Value        float64 `json:"value"`
	ScriptPubKey struct {
		Address string `json:"address"`
	} `json:"scriptPubKey"`
}

type RawTransaction struct {
	TXID string `json:"txid"`
	Vin  []Vin  `json:"vin"`
	Vout []Vout `json:"vout"`
}

func decodeTransaction(txid string, blockhash string) {
	var tx RawTransaction
	if err := callRPC("getrawtransaction", []interface{}{txid, true, blockhash}, &tx); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("\n--- Decoded Transaction ---")
	fmt.Printf("TXID: %s\n", tx.TXID)

	fmt.Println("Inputs:")
	for _, vin := range tx.Vin {
		if vin.Coinbase != "" {
			fmt.Println("  Coinbase (newly mined coins)")
		} else {
			fmt.Printf("  From TXID: %s, Vout: %d\n", vin.TXID, vin.Vout)
		}
	}

	fmt.Println("Outputs:")
	for _, vout := range tx.Vout {
		fmt.Printf("  %.8f BTC -> %s\n", vout.Value, vout.ScriptPubKey.Address)
	}
}

// Challenge 5: Block Details
type Block struct {
	Hash         string   `json:"hash"`
	Height       int      `json:"height"`
	Time         int64    `json:"time"`
	Nonce        uint32   `json:"nonce"`
	Difficulty   float64  `json:"difficulty"`
	PreviousHash string   `json:"previousblockhash"`
	Transactions []string `json:"tx"`
}

func getBlockDetails(blockhash string) {
	var block Block
	if err := callRPC("getblock", []interface{}{blockhash}, &block); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("\n--- Block Details ---")
	fmt.Printf("Hash:         %s\n", block.Hash)
	fmt.Printf("Height:       %d\n", block.Height)
	fmt.Printf("Time:         %d\n", block.Time)
	fmt.Printf("Nonce:        %d\n", block.Nonce)
	fmt.Printf("Difficulty:   %f\n", block.Difficulty)
	fmt.Printf("Previous:     %s\n", block.PreviousHash)
	fmt.Printf("Transactions: %d\n", len(block.Transactions))
	fmt.Println("TXIDs:")
	for _, txid := range block.Transactions {
		fmt.Printf("  %s\n", txid)
	}
}

func main() {
	fmt.Println("Bitcoin Explorer - Day 2")
	fmt.Println("========================")
	getBlockchainInfo()
	getWalletBalance()
	listTransactions()
	decodeTransaction("502f29ad32725c51ede6b37e8fe5fc03cce81489c964d9a509119ea3177f0cfe",
		"1df39d01f9d10d6b8db2ea73572422d7df1edfbf0af0d8d9e026ab846d9ce9ea",
	)
	getBlockDetails("1df39d01f9d10d6b8db2ea73572422d7df1edfbf0af0d8d9e026ab846d9ce9ea")
}

// clone lnd
