package blockchain

import (
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
)

const (
	dbFolder = ".blocks"
)

func openDB(path string) (*badger.DB, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}

	return badger.Open(opts)
}

func dbPath() string {
	return filepath.Join(dbFolder, "chaindata")
}

func (bc *Blockchain) CloseDB() error {
	return bc.DB.Close()
}
func (bc *Blockchain) Remove() {
	os.RemoveAll(dbPath())
}
