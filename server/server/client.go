package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

const secretByteLength = 32

// Client contains the information and hooks of a registered client
type Client struct {
	Name       string              `json:"name"`
	Secret     []byte              `json:"-"`
	CreatedAt  time.Time           `json:"createdAt"`
	LastAction time.Time           `json:"lastAction"`
	Hooks      map[string]*Webhook `json:"hooks"`
	ws         []*melody.Melody
}

func (c *Client) generateSecret() (string, error) {
	b := make([]byte, secretByteLength)
	n, err := rand.Read(b)

	if n != 32 {
		return "", &ErrSecretTooShort{n: n}
	}

	if err != nil {
		return "", err
	}

	s := c.Name + ":" + base64.URLEncoding.EncodeToString(b)
	c.Secret = sha256.New().Sum([]byte(s))

	return s, nil
}

// OpenWebsocket opens a socket for this client that listens to all hooks
func (c *Client) OpenWebsocket(con *gin.Context) {
	m := melody.New()
	if c.ws == nil {
		c.ws = make([]*melody.Melody, 0)
	}
	c.ws = append(c.ws, m)

	err := m.HandleRequest(con.Writer, con.Request)
	if err != nil {
		log.Print(err)
		err = m.CloseWithMsg([]byte(err.Error()))
		if err != nil {
			log.Print(err)
		}
		con.Status(http.StatusInternalServerError)
	}
}

// Destroy this client and all related webhooks and connections
func (c *Client) Destroy() {
	for _, w := range c.ws {
		w.Close()
	}
}

// MarshalJSON marshals a client to the correct json representation
func (c *Client) MarshalJSON() ([]byte, error) {
	_, h := hookMapToSlice(c.Hooks)
	cli := struct {
		Name       string     `json:"name"`
		CreatedAt  time.Time  `json:"createdAt"`
		LastAction time.Time  `json:"lastAction"`
		Hooks      []*Webhook `json:"hooks"`
	}{
		c.Name,
		c.CreatedAt,
		c.LastAction,
		h,
	}

	b, err := json.Marshal(cli)
	return b, err
}

// UnmarshalJSON unmarshals the JSON representation of a client
func (c *Client) UnmarshalJSON(in []byte) error {
	cli := struct {
		Name       string     `json:"name"`
		CreatedAt  time.Time  `json:"createdAt"`
		LastAction time.Time  `json:"lastAction"`
		Hooks      []*Webhook `json:"hooks"`
	}{}

	err := json.Unmarshal(in, &cli)
	if err != nil {
		return err
	}

	c.Name = cli.Name
	c.CreatedAt = cli.CreatedAt
	c.LastAction = cli.LastAction
	c.Hooks = make(map[string]*Webhook)
	for _, v := range cli.Hooks {
		c.Hooks[v.Identifier] = v
	}

	return nil
}

func hookMapToSlice(inmap map[string]*Webhook) ([]string, []*Webhook) {
	keys := make([]string, 0)
	values := make([]*Webhook, 0)
	for k, v := range inmap {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}
