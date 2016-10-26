package node

import (
	"encoding/json"
	"time"
	"bytes"
	"net/http"
	"io"
)

type state struct {
	Next string
	ID string
	Prev string
}


/* Periodically sends the nodes current state to the state server*/
func (n *Node) updateState() {

	client := &http.Client{}
	for {
		s := n.newState()
		n.updateReq(s, client)
		time.Sleep(time.Second * 1)
	}

}
/* Creates a new state */
func (n *Node) newState() io.Reader {
	s := state {
		Next: n.table.fingers[1].node.Ip,
		ID: n.Ip,
		Prev: n.prev.Ip }

	buff := new(bytes.Buffer)

	err := json.NewEncoder(buff).Encode(s)
	if err != nil {
		n.logger.Error(err.Error())
	}

	return bytes.NewReader(buff.Bytes())
}
/* Sends the node state to the state server*/
func (n *Node) updateReq(r io.Reader, c *http.Client) {
	req, err := http.NewRequest("POST", "http://129.242.22.74:8080/update", r)
	if err != nil {
		n.logger.Error(err.Error())
	}

	resp, err := c.Do(req)
	if err != nil {
		n.logger.Error(err.Error())
	} else {
		resp.Body.Close()
	}
}

/* Sends a post request to the state server add endpoint */
func (n *Node) add() {
	r := n.newState()
	req, err := http.NewRequest("POST", "http://129.242.22.74:8080/add", r)
	if err != nil {
		n.logger.Error(err.Error())
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		n.logger.Error(err.Error())
	} else {
		resp.Body.Close()
	}
}


