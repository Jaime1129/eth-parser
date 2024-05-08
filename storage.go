package main

import "sync"

type Storage interface {
	AddSubcriber(address string)
	GetSubsriberList() []string
	GetTrxsByAddress(address string) []Transaction
	SaveTrxs(trxs map[string][]Transaction)
}

type storage struct {
	lock           sync.RWMutex             // protect read or write of subscribers
	subscribedTrxs map[string][]Transaction // address => incoming or outgoing trx list
}

func NewStorage() Storage {
	return &storage{
		subscribedTrxs: make(map[string][]Transaction),
	}
}

func (s *storage) AddSubcriber(address string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.subscribedTrxs[address]; !ok {
		s.subscribedTrxs[address] = make([]Transaction, 0)
	}
}

func (s *storage) GetSubsriberList() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	res := make([]string, len(s.subscribedTrxs))
	for k, _ := range s.subscribedTrxs {
		res = append(res, k)
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
