package node

import (
	"encoding/json"
//	"time"
	"bytes"
	"net/http"
	"io"
	"fmt"
)

type state struct {
	Next string
	ID string
	Prev string
}


/* Periodically sends the nodes current state to the state server*/
func (n *Node) updateState() {

	client := &http.Client{}
//	for {
	s := n.newState()
	n.updateReq(s, client)
//		time.Sleep(time.Second * 1)

//	}

}
/* Creates a new state */
func (n *Node) newState() io.Reader {
	succ := n.getSuccessor()
	prev := n.getPrev()

	id := fmt.Sprintf("%s:%s", n.info.Ip, n.info.HttpPort)
	nextId := fmt.Sprintf("%s:%s", succ.Ip, succ.HttpPort)
	prevId := fmt.Sprintf("%s:%s", prev.Ip, prev.HttpPort)

	s := state {
		Next: nextId,
		ID: id,
		Prev: prevId }

	buff := new(bytes.Buffer)

	err := json.NewEncoder(buff).Encode(s)
	if err != nil {
		n.logger.Error(err.Error())
	}

	return bytes.NewReader(buff.Bytes())
}
/* Sends the node state to the state server*/
func (n *Node) updateReq(r io.Reader, c *http.Client) {
	req, err := http.NewRequest("POST", "http://129.242.22.74:7560/update", r)
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
	req, err := http.NewRequest("POST", "http://129.242.22.74:7560/add", r)
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

func (n *Node) remove() {
	r := n.newState()
	req, err := http.NewRequest("POST", "http://129.242.22.74:7560/remove", r)
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

