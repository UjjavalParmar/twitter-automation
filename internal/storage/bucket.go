package storage

import (
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type Store struct {
	db *badger.DB
}

func Open(dir string) (*Store, error) {
	opts := badger.DefaultOptions(filepath.Join(dir, "bot.db"))
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// MarkPosted stores a unique key per day+slot to avoid double posting.
func (s *Store) MarkPosted(key string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry([]byte("posted:"+key), []byte(time.Now().Format(time.RFC3339))))
	})
}

func (s *Store) WasPosted(key string) (bool, error) {
	var found bool
	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("posted:" + key))
		if err == badger.ErrKeyNotFound {
			found = false
			return nil
		}
		if err != nil {
			return err
		}
		found = true
		return nil
	})
	return found, err
}

func (s *Store) SeenTweet(id string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("seen:"+id), []byte("1"))
	})
}

func (s *Store) IsSeen(id string) (bool, error) {
	var seen bool
	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("seen:" + id))
		if err == badger.ErrKeyNotFound {
			seen = false
			return nil
		}
		if err != nil {
			return err
		}
		seen = true
		return nil
	})
	return seen, err
}
