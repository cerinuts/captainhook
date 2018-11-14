package server

import (
	"runtime"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

const defaultDatabasePath = "/var/cerinuts/captainhook/db"
const defaultDatabasePathWin = "./db"

const delimeter = "."

type DB struct {
	bdb  *badger.DB
	path string
}

func Open(path string) *DB {

	if path == "" {
		if runtime.GOOS == "windows" {
			path = defaultDatabasePathWin
		} else {
			path = defaultDatabasePath
		}
	}

	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	opts.Truncate = true
	opts.ValueLogFileSize = 1024 * 1024
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}

	return &DB{
		bdb:  db,
		path: path,
	}
}

func (db *DB) Store(client *Client) error {
	return db.bdb.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(client.Name+delimeter+"Secret"), []byte(client.Secret))
		if err != nil {
			return err
		}

		err = txn.Set([]byte(client.Name+delimeter+"CreatedAt"), []byte(client.CreatedAt.Format(time.RFC3339)))
		if err != nil {
			return err
		}

		err = txn.Set([]byte(client.Name+delimeter+"LastAction"), []byte(client.LastAction.Format(time.RFC3339)))
		if err != nil {
			return err
		}

		for _, h := range client.Hooks {
			err = txn.Set([]byte(client.Name+delimeter+"Hooks"+delimeter+h.Identifier+delimeter+"URL"), []byte(h.URL))
			if err != nil {
				return err
			}

			err = txn.Set([]byte(client.Name+delimeter+"Hooks"+delimeter+h.Identifier+delimeter+"UUID"), []byte(h.UUID))
			if err != nil {
				return err
			}

			err = txn.Set([]byte(client.Name+delimeter+"Hooks"+delimeter+h.Identifier+delimeter+"CreatedAt"), []byte(h.CreatedAt.Format(time.RFC3339)))
			if err != nil {
				return err
			}

			err = txn.Set([]byte(client.Name+delimeter+"Hooks"+delimeter+h.Identifier+delimeter+"LastCall"), []byte(h.LastCall.Format(time.RFC3339)))
			if err != nil {
				return err
			}

		}

		return err
	})
}

func (db *DB) Load() (map[string]*Client, error) {
	clients := make(map[string]*Client)
	err := db.bdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				return handleKeyValuePair(string(k), string(v), clients)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return clients, err
}

func (db *DB) Delete(clientName string) error {
	return db.bdb.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(clientName)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			err := txn.Delete(k)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func handleKeyValuePair(k, v string, clients map[string]*Client) error {
	keysplit := strings.Split(k, delimeter)
	name := keysplit[0]
	if clients[name] == nil {
		clients[name] = new(Client)
		clients[name].Hooks = make(map[string]*Webhook)
		clients[name].Name = name
	}
	switch keysplit[1] {
	case "Secret":
		clients[name].Secret = v
	case "CreatedAt":
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return err
		}
		clients[name].CreatedAt = t
	case "LastAction":
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return err
		}
		clients[name].LastAction = t
	case "Hooks":
		if clients[name].Hooks[keysplit[2]] == nil {
			clients[name].Hooks[keysplit[2]] = new(Webhook)
			clients[name].Hooks[keysplit[2]].Identifier = keysplit[2]
		}
		switch keysplit[3] {
		case "URL":
			clients[name].Hooks[keysplit[2]].URL = v
		case "UUID":
			clients[name].Hooks[keysplit[2]].UUID = v
		case "CreatedAt":
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return err
			}
			clients[name].Hooks[keysplit[2]].CreatedAt = t
		case "LastCall":
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return err
			}
			clients[name].Hooks[keysplit[2]].LastCall = t
		}
	}
	return nil
}
