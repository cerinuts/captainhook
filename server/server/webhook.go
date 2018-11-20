/*
Copyright (c) 2018 ceriath
This Package is part of "captainhook"
It is licensed under the MIT License
*/

package server

import (
	"bufio"
	"bytes"
	"net/http"
	"time"
)

// Webhook contains the information about a webhook
type Webhook struct {
	URL        string    `json:"url"`
	Identifier string    `json:"identifier"`
	UUID       string    `json:"uuid"`
	CreatedAt  time.Time `json:"createdAt"`
	LastCall   time.Time `json:"lastCall"`
	client     *Client
}

// Handle relays the request to all connected clients
func (w *Webhook) Handle(req *http.Request) error {
	w.LastCall = time.Now()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	err := req.WriteProxy(writer)
	if err != nil {
		log.Errorf("Cloud not write request: %s ", err.Error())
		return err
	}
	err = writer.Flush()
	if err != nil {
		log.Error("Error flushing writer")
		return err
	}

	for _, ws := range w.client.ws {
		err = ws.Broadcast(b.Bytes())
		if err != nil {
			log.Error("Could not send to websocket")
		}
	}

	return nil
}
