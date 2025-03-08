package state

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/karimseh/gochain/pkg/crypto"
	"github.com/karimseh/gochain/pkg/types"
)

type State struct {
	db    *badger.DB
	cache map[string]*types.Account
	mu    sync.RWMutex
}

func NewState(db *badger.DB) *State {
	return &State{
		db:    db,
		cache: make(map[string]*types.Account),
	}
}
func (s *State) GetBalance(address string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	acc, err := s.GetAccount(address)
	if err != nil {
		return 0, err
	}
	return acc.Balance, nil
}

func (s *State) GetAccount(address string) (*types.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if acc, exists := s.cache[address]; exists {
		return acc, nil
	}

	var acc *types.Account
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("account-" + address))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &acc)
		})
	})

	if err == badger.ErrKeyNotFound {
		return &types.Account{Address: address, Balance: 0, Nonce: 0}, nil
	}
	if err == nil {
		s.cache[address] = acc
	}
	return acc, err
}

func (s *State) SaveAccount(acc *types.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[acc.Address] = acc
	data, err := json.Marshal(acc)
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("account-"+acc.Address), data)
	})
}

func (s *State) ValidateTx(tx *types.Transaction) error {
	if !tx.Verify() {
		return fmt.Errorf("transaction signature verification failed")
	}

	sender, err := s.GetAccount(tx.From)
	if err != nil {
		return err
	}
	if tx.Ammount <= 0 {
		return fmt.Errorf("transaction ammount must be greater than 0")
	}

	if tx.Nonce != sender.Nonce+1 {
		return fmt.Errorf("invalid nonce: %d, expected: %d", tx.Nonce, sender.Nonce+1)
	}

	if sender.Balance < tx.Ammount {
		return fmt.Errorf("insufficient balance: %d, required: %d", sender.Balance, tx.Ammount)
	}
	return nil
}
func (s *State) ApplyTx(tx *types.Transaction) error {
	if err := s.ValidateTx(tx); err != nil {
		return err
	}
	sender, _ := s.GetAccount(tx.From)
	sender.Balance -= tx.Ammount
	sender.Nonce = tx.Nonce
	if err := s.SaveAccount(sender); err != nil {
		return err
	}

	reciever, _ := s.GetAccount(tx.To)
	reciever.Balance += tx.Ammount
	return s.SaveAccount(reciever)

}

func (s *State) ApplyBlock(block *types.Block) error {
	coinbase := block.Transactions[0]
	if err := s.ApplyCoinbase(coinbase); err != nil {
		return err
	}

	for _, tx := range block.Transactions[1:] {
		if err := s.ApplyTx(tx); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) ApplyCoinbase(tx *types.Transaction) error {
	acc, _ := s.GetAccount(tx.To)
	acc.Balance += tx.Ammount
	return s.SaveAccount(acc)
}

func (s *State) GetNextNonce(address string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	acc, err := s.GetAccount(address)
	if err != nil {
		return 0, err
	}
	return acc.Nonce + 1, nil
}

func (s *State) CalculateStateRoot() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all accounts from DB
	accounts, err := s.getAllAccounts()
	if err != nil {
		return nil, err
	}

	// Sort accounts by address for deterministic ordering
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Address < accounts[j].Address
	})

	hashes := make([][]byte, len(accounts))
	for i, acc := range accounts {
		data, err := serializeAccount(acc)
		if err != nil {
			return nil, err
		}
		hashes[i] = crypto.HashData(data)
	}

	return crypto.BuildMerkleRoot(hashes), nil
}

func (s *State) getAllAccounts() ([]*types.Account, error) {
	var accounts []*types.Account
	
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("account-")
		
		it := txn.NewIterator(opts)
		defer it.Close()
		
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var acc types.Account
				if err := json.Unmarshal(v, &acc); err != nil {
					return err
				}
				accounts = append(accounts, &acc)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	
	return accounts, err
}
func serializeAccount(acc *types.Account) ([]byte, error) {
	return json.Marshal(struct {
		Address string
		Balance uint64
		Nonce   uint64
	}{
		Address: acc.Address,
		Balance: acc.Balance,
		Nonce:   acc.Nonce,
	})
}
