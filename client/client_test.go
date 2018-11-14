package captainhook

import (
	"testing"
)

func TestClientConnect(t *testing.T) {
	tables := []struct {
		host   string
		port   string
		secret string
		useSSL bool
	}{
		{"localhost", "8080", "test:abc", false},
		{"localhost", "8082", "test:abc", true},
	}

	for _, table := range tables {
		cli, err := NewClient(table.secret, nil)
		if err != nil {
			t.Errorf("Error creating client %s", err.Error())
		}
		_, err = cli.Connect(table.host, table.port, table.useSSL)
		if err != nil {
			t.Errorf("Error connecting: %s", err.Error())
		}
		err = cli.Disconnect()
		if err != nil {
			t.Errorf("Error disconnecting: %s", err.Error())
		}
	}
}

func TestClientAddHook(t *testing.T) {
	tables := []struct {
		host       string
		port       string
		secret     string
		identifier string
		useSSL     bool
	}{
		{"localhost", "8082", "test:abc", "abc", true},
		{"localhost", "8080", "test:abc", "abc", false},
	}

	for _, table := range tables {
		cli, err := NewClient(table.secret, nil)
		if err != nil {
			t.Errorf("Error creating client %s", err.Error())
		}
		_, err = cli.Connect(table.host, table.port, table.useSSL)
		if err != nil {
			t.Errorf("Error connecting: %s", err.Error())
		}

		err = cli.AddHook(table.identifier)
		if err != nil {
			t.Errorf("Error creating hook: %s", err.Error())
		}
		err = cli.Disconnect()
		if err != nil {
			t.Errorf("Error disconnecting: %s", err.Error())
		}
	}
}

func TestClientRemoveHook(t *testing.T) {
	tables := []struct {
		host       string
		port       string
		secret     string
		identifier string
		useSSL     bool
	}{
		{"localhost", "8082", "test:abc", "abcx", true},
		{"localhost", "8080", "test:abc", "abcx", false},
	}

	for _, table := range tables {
		cli, err := NewClient(table.secret, nil)
		if err != nil {
			t.Errorf("Error creating client %s", err.Error())
		}
		_, err = cli.Connect(table.host, table.port, table.useSSL)
		if err != nil {
			t.Errorf("Error connecting: %s", err.Error())
		}

		err = cli.AddHook(table.identifier)
		if err != nil {
			t.Errorf("Error creating hook: %s", err.Error())
		}

		err = cli.RemoveHook(table.identifier)
		if err != nil {
			t.Errorf("Error removing hook: %s", err.Error())
		}
		err = cli.Disconnect()
		if err != nil {
			t.Errorf("Error disconnecting: %s", err.Error())
		}
	}
}
