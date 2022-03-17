package database

import (
	"fmt"
	"os"

	// "path/filepath"
	"time"
)

type State struct {
	Balances  map[AccountAddress]uint
	TxMempool TransactionList
	dbFile    *os.File

	lastBlockSerialNo  int
	lastBlockTimestamp int64
	latestHash         [32]byte
	latestTimestamp    int64
}

func makeTimestamp() int64 {
	return time.Now().UnixNano()
}

func (s *State) getNextBlockSerialNo() int {
	return s.lastBlockSerialNo + 1
}

func (s *State) getLatestHash() [32]byte {
	return s.latestHash
}

func LoadState() (*State, error) {
	var file *os.File
	state := &State{make(map[AccountAddress]uint), make([]Transaction, 0), file, 0, 0, [32]byte{}, 0}

	// loadedTransactions := LoadTransactions()
	// for _, t := range loadedTransactions {
	// 	if err := state.AddTransaction(t); err != nil {
	// 		panic("Transaction not allowed\n\t" + err.Error())
	// 	}
	// }

	localBlockchain := LoadBlockchain()
	err := state.ApplyBlocks(localBlockchain)
	if err != nil {
		panic(err)
	}

	return state, nil
}

func (state *State) AddTransaction(transaction Transaction) error {
	if err := state.ValidateTransaction(transaction); err != nil {
		return err
	}

	state.TxMempool = append(state.TxMempool, transaction)

	state.ApplyTransaction(transaction)

	state.latestTimestamp = transaction.Timestamp
	return nil
}

func (state *State) ApplyTransaction(transaction Transaction) {
	if transaction.Type != "genesis" && transaction.Type != "reward" {
		state.Balances[transaction.From] -= uint(transaction.Amount)
	}
	state.Balances[transaction.To] += uint(transaction.Amount)
}

func (state *State) ValidateTransaction(transaction Transaction) error {
	if (state.lastBlockSerialNo == 0 && transaction.Type == "genesis") || transaction.Type == "reward" {
		return nil
	}

	if transaction.From == transaction.To {
		return fmt.Errorf("a normal transaction is not allowed to same account")
	}

	if _, ok := state.Balances[transaction.From]; !ok {
		return fmt.Errorf("sending from Undefined Account")
	}
	if transaction.Amount <= 0 {
		return fmt.Errorf("illegal to make a transaction with 0 or less coins")
	}

	if transaction.Timestamp <= state.latestTimestamp {
		return fmt.Errorf("new tx must have newer timestamp than previous tx")
	}

	if state.Balances[transaction.From] < uint(transaction.Amount) {
		return fmt.Errorf("u broke")
	}

	return nil
}

func (state *State) ValidateTransactionList(transactionList TransactionList) error {
	for i, t := range transactionList {
		err := state.ValidateTransaction(t)
		if err != nil {
			return fmt.Errorf("Transaction nr. %d is not valid. Received Error: %s", i, err.Error())
		}
	}
	return nil
}

func (state *State) AddTransactionList(transactionList TransactionList) error {
	for i, t := range transactionList {
		err := state.AddTransaction(t)
		if err != nil {
			return fmt.Errorf("Transaction nr. %d is not able to be added. Received Error: %s", i, err.Error())
		}
	}
	return nil
}
