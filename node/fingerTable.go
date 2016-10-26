package node

import (
	"github.com/joonnna/ds_chord/node_communication"
	"github.com/joonnna/ds_chord/util"
	"math/rand"
	"math/big"
	"time"
	"errors"
	"net/http"
)

const (
	lenOfId = 160
)


type fingerEntry struct {
	node shared.NodeInfo
	start big.Int
}

type fingerTable struct {
	fingers []fingerEntry
}


var (
	ErrFingerNotFound = errors.New("Cant find closesetpreceding finger")
)

/* Calculates the start of the given interval
	formula: (n+2^k-1) mod 2^m
	exponent: The exponent k of the given formula
	modExp: The exponent m of the given formula
	id: n in the given formula
	Returns the start of the interval
*/
func calcStart(exponent int, modExp int, id big.Int) big.Int {
	base2 := big.NewInt(int64(2))
	k := big.NewInt(int64(exponent))

	tmp := big.NewInt(int64(0))
	tmp.Exp(base2, k, nil)

	sum := big.NewInt(int64(0))
	sum.Add(tmp, &id)

	modExponent := big.NewInt(int64(modExp))
	mod := big.NewInt(int64(0))
	mod.Exp(base2, modExponent, nil)

	ret := big.NewInt(int64(0))

	ret.Mod(sum, mod)

	return *ret
}

/* Inits the finger table, calcluates the start of each interval*/
func (n *Node) initFingerTable() {
	n.table.fingers = make([]fingerEntry, lenOfId)

	for i := 1; i < (lenOfId-1); i++ {
		n.table.fingers[i].start = calcStart((i-1), lenOfId, n.id)
	}
}
/* Periodically updates the fingert table by finding the successor
   node of each interval start in the fingertable.*/
func (n *Node) fixFingers() {
	for {

		/* Alone in the ring, no need to update table */
		if n.table.fingers[1].node.Ip == n.Ip {
			time.Sleep(time.Second * 1)
			continue
		}

		/* Index 1 is the successor and index 0 is not used */
		index := rand.Int() % lenOfId
		if index == 1 || index == 0 {
			index = 2
		}

		node, err := n.findPreDecessor(n.table.fingers[index].start)
		if err != nil {
			n.logger.Error(err.Error())
			time.Sleep(time.Second*1)
			continue
		}

		succ, err := n.getSucc(node.Ip)
		if err != nil {
			time.Sleep(time.Second*1)
			continue
		}
		n.table.fingers[index].node = succ
		time.Sleep(time.Second*1)
	}
}
/* Periodically called to stabilize ring position
   Queries the successor for its predeccessor and notifies it of
   the current nodes existence
*/
func(n *Node) stabilize() {
	var succ string

	arg := 0

	succ = n.table.fingers[1].node.Ip

	r := &shared.Search{}
	if succ == n.Ip {
		return
	}
	/* Can't exceed limit of idle connections */
	http.DefaultTransport.(*http.Transport).CloseIdleConnections()

	err := shared.SingleCall("Node.GetPreDecessor", succ, n.RpcPort, arg, r)
	if err != nil {
		n.logger.Error(err.Error())
		return
	}

	tmp := new(big.Int)
	tmp.SetBytes(r.Id)

	if util.BetweenNodes(n.id, n.table.fingers[1].node.Id, *tmp) {
		n.table.fingers[1].node.Ip = r.Ip
		n.table.fingers[1].node.Id = *tmp
	}

	args := shared.Search {
		Ip: n.Ip,
		Id: n.id.Bytes()}

	reply := &shared.Reply{}
	err = shared.SingleCall("Node.Notify", n.table.fingers[1].node.Ip, n.RpcPort, args, reply)
	if err != nil {
		n.logger.Error(err.Error())
	}
}
/* Joins the network
   Alerts nameserver of its presence
   Gets list of all nodes from nameserver
   Finds successor and inserts itself into the network
   Predecessor is initially set to itself.
*/
func (n *Node) join(nodeIp string) {

	/* Alone, set successor and predecessor to myself */
	if nodeIp == "" {
		self := shared.NodeInfo {
			Id: n.id,
			Ip: n.Ip }
		n.table.fingers[1].node = self
		n.prev = self
		return
	}
	node := shared.Search {
		Ip: n.Ip,
		Id: n.id.Bytes() }

	r := &shared.Reply{}

	err = shared.SingleCall("Node.FindSuccessor", randomNode, n.RpcPort, node, r)
	if err != nil {
		n.logger.Error(err.Error())
		return
	}

	n.prev.Id = n.id
	n.prev.Ip = n.Ip
	n.table.fingers[1].node = r.Next
}
/* Local function for closest preceding finger
   Finds a node in the fingertable which is in the keyspace
   between own id and the given id.

   id: search key

   returns information about the found node.
*/
func (n *Node) closestFinger(id big.Int) shared.NodeInfo {
	for i := (lenOfId-1); i >= 1; i-- {
		entry := n.table.fingers[i].node.Id
		if entry.BitLen() != 0 && util.BetweenNodes(n.id, id, entry) {
			return n.table.fingers[i].node
		}
	}
	self := shared.NodeInfo {
		Ip: n.Ip,
		Id: n.id }

	return self
}
/* Wrapper for GetSuccessor
   Gets the successor of the given node.

   ip: ip address of the node to query

   Returns the successor of the given node
*/
func (n *Node) getSucc(ip string) (shared.NodeInfo, error) {
	var err error
	args := 0
	r := &shared.Search{}

	if (ip == n.Ip) {
		return n.table.fingers[1].node, nil
	} else {
		err = shared.SingleCall("Node.GetSuccessor", ip, n.RpcPort, args, r)
	}

	tmp := new(big.Int)
	tmp.SetBytes(r.Id)
	retVal := shared.NodeInfo {
		Ip: r.Ip,
		Id: *tmp }
	return retVal, err
}
/* Finds the predecessor of the given node
   Searches local fingertables first, then searches other nodes
   fingertables by using rpc

   id: search key

   Returns the predecessor of the given node
*/
func (n *Node) findPreDecessor(id big.Int) (shared.NodeInfo, error) {
	var err error
	self := shared.NodeInfo {
		Ip: n.Ip,
		Id: n.id }
	currNode := self
	succ := n.table.fingers[1].node

	r := &shared.Reply{}
	args := shared.Search {
		Id: id.Bytes() }

	for {
		/* In my own keyspace, no need to search more */
		if util.InKeySpace(currNode.Id, succ.Id, id) {
			break
		}

		/* Local search or remot search */
		if currNode.Ip == n.Ip {
			currNode = n.closestFinger(id)
		} else {
			err = shared.SingleCall("Node.ClosestPrecedingFinger", currNode.Ip, n.RpcPort, args, r)
			if err != nil {
				return currNode, err
			}
			currNode = r.Next
		}

		/* Need to get the succesor of the current node to check its keyspace */
		if currNode.Ip == n.Ip {
			succ = n.table.fingers[1].node
		} else {
			succ, err = n.getSucc(currNode.Ip)
			if err != nil {
				return currNode, err
			}
		}
	}

	return currNode, nil
}

