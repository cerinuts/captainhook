package captainhook

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"code.cerinuts.io/cerinuts/captainhook/server/server"
	"github.com/gorilla/websocket"
)

// Client contains all the information needed to connect to the CaptainHook server
type Client struct {
	secret                       string
	rootCAs                      *x509.CertPool
	ws                           *websocket.Conn
	Receiver                     chan *http.Request
	host, port, scheme, wsscheme string
}

// NewClient creates a new CaptainHook client. Use the secret you received from your server administrator. rootCAs can contain additional certifactes for SSL validation,
// e.g. snakeoil/self-signed. Set to nil if you don't expect that
func NewClient(secret string, rootCAs *x509.CertPool) (Client, error) {

	if rootCAs == nil {
		var err error
		rootCAs, err = x509.SystemCertPool()
		if err != nil {
			return Client{}, err
		}
	}

	client := Client{
		secret:   secret,
		rootCAs:  rootCAs,
		Receiver: make(chan *http.Request),
	}

	return client, nil
}

// Connect to the captainhook server. Provide the hostname and port of the server. If the server offers an SSL connection, you should set useSSL to true.
func (c *Client) Connect(host, port string, useSSL bool) (*http.Response, error) {
	c.host = host
	c.port = port
	if useSSL {
		c.scheme = "https"
		c.wsscheme = "wss"
	} else {
		c.scheme = "http"
		c.wsscheme = "ws"
	}

	u := url.URL{Scheme: c.wsscheme, Host: c.host + ":" + c.port, Path: server.ConnectPath}

	header := http.Header{
		"Authorization": []string{"Bearer " + c.secret},
	}

	dialer := &websocket.Dialer{TLSClientConfig: &tls.Config{RootCAs: c.rootCAs}}

	ws, resp, err := dialer.Dial(u.String(), header)
	c.ws = ws
	if err != nil {
		if resp != nil {
			if resp.StatusCode == 307 {
				var location *url.URL
				location, err = resp.Location()
				if err != nil {
					return resp, &server.ErrInvalidLocation{Message: err.Error()}
				}
				return c.Connect(location.Hostname(), location.Port(), true)
			}
			return nil, &server.ErrCouldNotConnect{Message: strconv.Itoa(resp.StatusCode)}
		}
		return resp, err
	}

	go func() {
		for {
			_, message, err := c.ws.ReadMessage()
			if err != nil {
				return
			}
			r := bufio.NewReader(bytes.NewReader(message))
			req, err := http.ReadRequest(r)
			if err != nil {
				return
			}
			c.Receiver <- req
		}
	}()
	return nil, nil
}

// AddHook will add a new Webhook to the server identified by identifier.
func (c *Client) AddHook(identifier string) error {
	u := url.URL{Scheme: c.scheme, Host: c.host + ":" + c.port, Path: server.HookPath + "/" + identifier}

	req := &http.Request{
		Method: "PUT",
		URL:    &u,
		Header: http.Header{
			"Authorization": []string{"Bearer " + c.secret},
		},
	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: c.rootCAs},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return &server.ErrUnknownServerError{Message: strconv.Itoa(resp.StatusCode)}
	}

	return nil
}

// RemoveHook deletes the Webhook identified by the identifier on the server.
func (c *Client) RemoveHook(identifier string) error {
	u := url.URL{Scheme: c.scheme, Host: c.host + ":" + c.port, Path: server.HookPath + "/" + identifier}

	req := &http.Request{
		Method: "DELETE",
		URL:    &u,
		Header: http.Header{
			"Authorization": []string{"Bearer " + c.secret},
		},
	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: c.rootCAs},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return &server.ErrUnknownServerError{Message: strconv.Itoa(resp.StatusCode)}
	}

	return nil
}

// Disconnect disconnects the websocket from the server
func (c *Client) Disconnect() error {
	err := c.ws.Close()
	return err
}
