/*
Copyright (c) 2018 ceriath
This Package is part of "captainhook"
It is licensed under the MIT License
*/

package server

import (
	"crypto/sha256"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

// ApplicationName is the name of the application
const ApplicationName = "CaptainHook"

// VersionMajor is the major version according to semver
const VersionMajor = "0"

// VersionMinor is the minor version according to semver
const VersionMinor = "1"

// VersionPatch is the patch according to semver
const VersionPatch = "0"

// FullVersion is the full semantic version
const FullVersion = VersionMajor + "." + VersionMinor + "." + VersionPatch

// Server contains all the information about clients and webhooks
type Server struct {
	Clients        map[string]*Client
	Hooks          map[string]*Webhook
	DB             *DB
	hostname, port string
}

// NewServer creates a new CaptainHook Server
func NewServer(host, port string) *Server {

	return &Server{
		Clients:  make(map[string]*Client),
		Hooks:    make(map[string]*Webhook),
		DB:       Open(""),
		hostname: host,
		port:     port,
	}
}

// Load loads the initial database content
func (s *Server) Load() {
	cli, err := s.DB.Load()
	if err != nil {
		log.Fatalf("Error reading database: %s", err.Error())
	}
	s.Clients = cli
}

// Stop stops the server
func (s *Server) Stop() {
	for _, c := range s.Clients {
		for _, ws := range c.ws {
			ws.Close()
		}
	}
	s.DB.bdb.Close()
}

// Run blocks endlessly
func (s *Server) Run() {
	defer s.Stop()
	c := make(chan int)
	<-c
}

// AddClient adds a new client to the server
func (s *Server) AddClient(name string) (string, error) {

	if strings.Contains(name, delimeter) {
		err := &ErrInvalidClientName{Name: name}
		log.Error(err)
		return "", err
	}

	if s.Clients[name] != nil {
		err := &ErrClientAlreadyExists{Name: name}
		log.Error(err)
		return "", err
	}

	c := &Client{
		Name:       name,
		CreatedAt:  time.Now(),
		LastAction: time.Now(),
		Hooks:      make(map[string]*Webhook),
	}

	secret, err := c.generateSecret()
	if err != nil {
		log.Error(err)
		return "", &ErrSecretGenerationFailed{Message: err.Error()}
	}

	s.Clients[name] = c
	s.DB.Store(c)

	return secret, nil
}

// RemoveClient will delete a client
func (s *Server) RemoveClient(name string) error {
	if s.Clients[name] == nil {
		err := &ErrClientNotExists{Name: name}
		log.Error(err)
		return err
	}

	s.Clients[name].Destroy()

	s.DB.Delete(name)

	delete(s.Clients, name)

	return nil
}

// AddHook will add a hook identified by identifier to the given client
func (s *Server) AddHook(clientname, identifier string) (*Webhook, error) {

	if s.Clients[clientname] == nil {
		err := &ErrClientNotExists{Name: clientname}
		log.Error(err)
		return nil, err
	}

	if s.Clients[clientname].Hooks[identifier] != nil {
		err := &ErrHookAlreadyExists{Identifier: identifier}
		log.Error(err)
		return nil, err
	}

	url, uuid, err := s.generateURL()
	if err != nil {
		log.Error(err)
		return nil, &ErrCreatingUUID{Message: err.Error()}
	}

	w := &Webhook{
		CreatedAt:  time.Now(),
		LastCall:   time.Now(),
		Identifier: identifier,
		URL:        url,
		UUID:       uuid,
		client:     s.Clients[clientname],
	}

	s.Hooks[uuid] = w
	s.Clients[clientname].Hooks[identifier] = w
	s.DB.Store(s.Clients[clientname])

	return w, nil
}

// DeleteHook removes the webhook identified by identifier from the given client
func (s *Server) DeleteHook(clientname, identifier string) error {
	if s.Clients[clientname] == nil {
		err := &ErrClientNotExists{Name: clientname}
		log.Error(err)
		return err
	}

	if s.Clients[clientname].Hooks[identifier] == nil {
		err := &ErrHookNotExists{Identifier: identifier}
		log.Error(err)
		return err
	}

	delete(s.Hooks, s.Clients[clientname].Hooks[identifier].UUID)
	delete(s.Clients[clientname].Hooks, identifier)

	return nil
}

// DeleteHookByUUID will remove the webhook identified by the uuid from the CaptainHook instance
func (s *Server) DeleteHookByUUID(uuid string) error {
	for _, c := range s.Clients {
		for _, h := range c.Hooks {
			if h.UUID == uuid {
				delete(c.Hooks, h.Identifier)
				delete(s.Hooks, h.UUID)
				s.DB.Store(c)
				return nil
			}

		}

	}

	err := &ErrHookNotExists{Identifier: uuid}
	log.Error(err)
	return err
}

// HandleHook will proxy the http request sent by the 3rd party to the client this webhook belongs to
func (s *Server) HandleHook(uuid string, req *http.Request) error {
	if s.Hooks[uuid] == nil {
		return &ErrHookNotExists{Identifier: uuid}
	}

	err := s.Hooks[uuid].Handle(req)
	if err != nil {
		log.Error(err)
		return err
	}

	//persist LastCall for webhook
	return s.DB.Store(s.Hooks[uuid].client)
}

// RegenerateClientSecret will recreate a secret for the given client and invalidate the old one
func (s *Server) RegenerateClientSecret(clientname string) (string, error) {
	if s.Clients[clientname] == nil {
		err := &ErrClientNotExists{Name: clientname}
		log.Error(err)
		return "", err
	}

	secret, err := s.Clients[clientname].generateSecret()
	if err != nil {
		log.Error(err)
		return "", err
	}

	err = s.DB.Store(s.Clients[clientname])
	if err != nil {
		log.Error(err)
		return "", err
	}

	return secret, nil
}

func (s *Server) validateClient(secret string) *Client {
	split := strings.Split(secret, ":")
	if len(split) != 2 {
		log.Infof("Clientsecretsplit length %d invalid", len(split))
		return nil
	}

	if s.Clients[split[0]] == nil {
		log.Infof("Client '%s' does not exist", split[0])
		return nil
	}

	sec := s.Clients[split[0]].Secret
	_ = sec
	if string(s.Clients[split[0]].Secret) != string(sha256.New().Sum([]byte(secret))) {
		log.Info("Clientsecret does not match")
		return nil
	}

	return s.Clients[split[0]]
}

// generateURL returns fullUrl (https://host:port/h/UUID), UUID
func (s *Server) generateURL() (string, string, error) {

	u4, err := uuid.NewV4()
	if err != nil {
		log.Error(err)
		return "", "", err
	}
	return "http://" + s.hostname + ":" + s.port + "/h/" + u4.String(), u4.String(), nil
}
