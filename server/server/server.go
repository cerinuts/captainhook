package server

import (
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

func (s *Server) Load() {
	cli, err := s.DB.Load()
	if err != nil {
		panic("Error reading database")
	}
	s.Clients = cli
}

func (s *Server) Stop() {
	for _, c := range s.Clients {
		for _, ws := range c.ws {
			ws.Close()
		}
	}
	s.DB.bdb.Close()
}

func (s *Server) Run() {
	defer s.Stop()
	c := make(chan int)
	<-c
}

// AddClient adds a new client to the server
func (s *Server) AddClient(name string) (string, error) {

	if strings.Contains(name, delimeter) {
		return "", &ErrInvalidClientName{Name: name}
	}

	if s.Clients[name] != nil {
		return "", &ErrClientExists{Name: name}
	}

	c := &Client{
		Name:       name,
		CreatedAt:  time.Now(),
		LastAction: time.Now(),
		Hooks:      make(map[string]*Webhook),
	}

	secret, err := c.generateSecret()
	if err == nil {
		return "", &ErrSecretGenerationFailed{Message: err.Error()}
	}

	s.Clients[name] = c
	s.DB.Store(c)

	return secret, nil
}

// RemoveClient will delete a client
func (s *Server) RemoveClient(name string) error {
	if s.Clients[name] == nil {
		return &ErrClientNotExists{Name: name}
	}

	s.Clients[name].Destroy()

	s.DB.Delete(name)

	delete(s.Clients, name)

	return nil
}

// AddHook will add a hook identified by identifier to the given client
func (s *Server) AddHook(clientname, identifier string) (*Webhook, error) {

	if s.Clients[clientname] == nil {
		return nil, &ErrClientNotExists{Name: clientname}
	}

	url, uuid, err := s.generateURL()
	if err != nil {
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
		return &ErrClientNotExists{Name: clientname}
	}

	if s.Clients[clientname].Hooks[identifier] == nil {
		return &ErrHookNotExists{Identifier: identifier}
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

	return &ErrHookNotExists{Identifier: uuid}
}

// HandleHook will proxy the http request sent by the 3rd party to the client this webhook belongs to
func (s *Server) HandleHook(uuid string, req *http.Request) error {
	if s.Hooks[uuid] == nil {
		return &ErrHookNotExists{Identifier: uuid}
	}

	err := s.Hooks[uuid].Handle(req)
	if err != nil {
		return err
	}

	//persist LastCall for webhook
	return s.DB.Store(s.Hooks[uuid].client)
}

func (s *Server) validateClient(secret string) *Client {
	split := strings.Split(secret, ":")
	if len(split) != 2 {
		return nil
	}

	if s.Clients[split[0]] == nil {
		return nil
	}

	sec := s.Clients[split[0]].Secret
	_ = sec
	if s.Clients[split[0]].Secret != secret {
		return nil
	}

	return s.Clients[split[0]]
}

// generateURL returns fullUrl (https://host:port/h/UUID), UUID
func (s *Server) generateURL() (string, string, error) {

	u4, err := uuid.NewV4()
	if err != nil {
		return "", "", err
	}
	return "http://" + s.hostname + ":" + s.port + "/h/" + u4.String(), u4.String(), nil
}
