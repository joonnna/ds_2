package client

import (
	"encoding/json"
//	"time"
	"bytes"
	"net/http"
	"io"
//	"fmt"
)

type state struct {
	Next string
	ID string
	Prev string
}


/* Periodically sends the nodes current state to the state server*/
func (c *Client) updateState(id, next, prev string) {
	client := &http.Client{}

	s := c.newState(id, next, prev)

	c.updateReq(s, client)

}
/* Creates a new state */
func (c *Client) newState(id, next, prev string) io.Reader {
	s := state {
		Next: next,
		ID: id,
		Prev: prev }

	buff := new(bytes.Buffer)

	err := json.NewEncoder(buff).Encode(s)
	if err != nil {
		c.log.Error(err.Error())
	}

	return bytes.NewReader(buff.Bytes())
}
/* Sends the node state to the state server*/
func (c *Client) updateReq(r io.Reader, httpClient *http.Client) {
	req, err := http.NewRequest("POST", "http://129.242.22.74:7560/update", r)
	if err != nil {
		c.log.Error(err.Error())
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		c.log.Error(err.Error())
	} else {
		resp.Body.Close()
	}
}

/* Sends a post request to the state server add endpoint */
/*
func (c *Client) add() {
	r := c.newState()
	req, err := http.NewRequest("POST", "http://129.242.22.74:7560/add", r)
	if err != nil {
		c.log.Error(err.Error())
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		c.log.Error(err.Error())
	} else {
		resp.Body.Close()
	}
}

func (c *Client) remove() {
	r := c.newState()
	req, err := http.NewRequest("POST", "http://129.242.22.74:7560/remove", r)
	if err != nil {
		c.log.Error(err.Error())
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		c.log.Error(err.Error())
	} else {
		resp.Body.Close()
	}
}
*/
