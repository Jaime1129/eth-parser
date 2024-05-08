package main

import (
	"sync"
	"sync/atomic"
)

type Storage interface {
	GetLatestBlock() int64
	SetLatestBlock(blockNum int64)
	AddSubcriber(address string)
	GetSubsriberAddresses() map[string]struct{}
	GetTrxsByAddress(address string) []Transaction
	SaveTrxs(trxs map[string][]Transaction)
}

type storage struct {
	lock           sync.RWMutex             // protect read or write of subscribers
	subscribedTrxs map[string][]Transaction // address => incoming or outgoing trx list
	blockNum       atomic.Int64
}

func NewStorage() Storage {
	return &storage{
		subscribedTrxs: make(map[string][]Transaction),
	}
}

func (s *storage) GetLatestBlock() int64 {
	return s.blockNum.Load()
}

func (s *storage) SetLatestBlock(blockNum int64) {
	s.blockNum.Store(blockNum)
}

func (s *storage) AddSubcriber(address string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.subscribedTrxs[address]; !ok {
		s.subscribedTrxs[address] = make([]Transaction, 0)
	}
}

func (s *storage) GetSubsriberAddresses() map[string]struct{} {
	s.lock.RLock()
	defer s.lock.RUnlock()

	res := make(map[string]struct{}, len(s.subscribedTrxs))
	for k, _ := range s.subscribedTrxs {
		res[k] = struct{}{}
	}

	return res
}

func (s *storage) GetTrxsByAddress(address string) []Transaction {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.subscribedTrxs[address]
}

func (s *storage) SaveTrxs(trxs map[string][]Transaction) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for k, v := range trxs {
		s.subscribedTrxs[k] = v
	}
}
