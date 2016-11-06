package node

import (
	"github.com/joonnna/ds_2/node_communication"
	"github.com/joonnna/ds_2/util"
	"math/rand"
	"math/big"
	"time"
	"errors"
	"fmt"
	"net/http"
	"strings"
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
		n.table.fingers[i].start = calcStart((i-1), lenOfId, n.info.Id)
	}
}
/* Periodically updates the fingert table by finding the successor
   node of each interval start in the fingertable.*/
func (n *Node) fixFingers() {

	time.Sleep(time.Second*1)
	/* Alone in the ring, no need to update table */
	curr_succ := n.getSuccessor()

	if equal := compareAddr(curr_succ, n.info); equal {
		return
	}

	/* Index 1 is the successor and index 0 is not used */
	index := rand.Int() % lenOfId
	if index == 1 || index == 0 {
		index = 2
	}

	node, err := n.findPreDecessor(n.table.fingers[index].start)
	if err != nil {
		n.logger.Error(err.Error())
		return
	}

	succ, err := n.getSucc(node)
	if err != nil {
		return
	}
	n.table.fingers[index].node = succ

}
/* Periodically called to stabilize ring position
   Queries the successor for its predeccessor and notifies it of
   the current nodes existence
*/
func(n *Node) stabilize() {
	var err error
	arg := 0

	r := &shared.Search{}

	reply := &shared.Reply{}

	for {

		dead_succ := false
		time.Sleep(time.Millisecond*500)

		firstSucc := n.getSuccessor()

		if equal := compareAddr(firstSucc, n.info); equal {
			continue
		}
		/* Can't exceed limit of idle connections */
		http.DefaultTransport.(*http.Transport).CloseIdleConnections()

		len := n.getSuccListLen()
		for i := 0; i < len; i++ {
			succ := n.getSuccIndex(i)
			err = shared.SingleCall("Node.GetPreDecessor", succ.Ip, succ.RpcPort, arg, r)
			if err == nil {
				if i > 0 {
					dead_succ = true
					n.clearSuccIndex(i)
					n.setSuccessor(succ)
				}
				break
			}
		}

		if err != nil {
			n.logger.Error("cant contact anyone")
			continue
		}

		tmp := new(big.Int)
		tmp.SetBytes(r.Id)

		curr_succ := n.getSuccessor()
		if util.BetweenNodes(n.info.Id, curr_succ.Id, *tmp) && !dead_succ{
			node := shared.NodeInfo {
				Ip: r.Ip,
				Id: *tmp,
				RpcPort: r.RpcPort,
				HttpPort: r.HttpPort }
			n.setSuccessor(node)
		}

		args := shared.Search {
			Ip: n.info.Ip,
			Id: n.info.Id.Bytes(),
			RpcPort: n.info.RpcPort,
			HttpPort: n.info.HttpPort }


		newSucc := n.getSuccessor()
		err = shared.SingleCall("Node.Notify", newSucc.Ip, newSucc.RpcPort, args, reply)
		if err != nil {
			n.logger.Error(err.Error())
		}

		n.updateSuccList()

		if dead_succ == false {
			n.fixFingers()
		}
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
		n.setSuccessor(n.info)
		n.setPrev(n.info)
		return
	}
	node := shared.Search {
		Ip: n.info.Ip,
		Id: n.info.Id.Bytes(),
		RpcPort: n.info.RpcPort,
		HttpPort: n.info.HttpPort }

	r := &shared.Reply{}

	addr := strings.Split(nodeIp, ":")
	err := shared.SingleCall("Node.FindSuccessor", addr[0], addr[1], node, r)
	if err != nil {
		n.logger.Error(err.Error())
		return
	}

	n.setPrev(r.Prev)
	n.setSuccessor(r.Next)
}
/* Local function for closest preceding finger
   Finds a node in the fingertable which is in the keyspace
   between own Id and the given Id.

   id: search key

   returns information about the found node.
*/
func (n *Node) closestFinger(id big.Int) shared.NodeInfo {
	for i := (lenOfId-1); i >= 1; i-- {
		entry := n.table.fingers[i].node.Id
		if entry.BitLen() != 0 && util.BetweenNodes(n.info.Id, id, entry) {
			return n.table.fingers[i].node
		}
	}

	return n.info
}
/* Wrapper for GetSuccessor
   Gets the successor of the given node.

   ip: ip address of the node to query

   Returns the successor of the given node
*/
func (n *Node) getSucc(node shared.NodeInfo) (shared.NodeInfo, error) {
	var err error
	args := 0
	r := &shared.Search{}

	if equal := compareAddr(node, n.info); equal {
		return n.getSuccessor(), nil
	} else {
		err = shared.SingleCall("Node.GetSuccessor", node.Ip, node.RpcPort, args, r)
	}

	tmp := new(big.Int)
	tmp.SetBytes(r.Id)
	retVal := shared.NodeInfo {
		Ip: r.Ip,
		Id: *tmp,
		RpcPort: r.RpcPort,
		HttpPort: r.HttpPort }

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

	currNode := n.info
	succ := n.getSuccessor()

	r := &shared.Reply{}
	args := shared.Search {
		Id: id.Bytes() }

	for {
		/* In my own keyspace, no need to search more */
		if util.InKeySpace(currNode.Id, succ.Id, id) {
			break
		}

		/* Local search or remote search */
		if equal := compareAddr(currNode, n.info); equal {
			currNode = n.closestFinger(id)
		} else {
			err = shared.SingleCall("Node.ClosestPrecedingFinger", currNode.Ip, currNode.RpcPort, args, r)
			if err != nil {
				return currNode, err
			}
			currNode = r.Next
		}

		/* Need to get the succesor of the current node to check its keyspace */
		if equal := compareAddr(currNode, n.info); equal {
			succ = n.getSuccessor()
		} else {
			succ, err = n.getSucc(currNode)
			if err != nil {
				return currNode, err
			}
		}
	}

	return currNode, nil
}


/* Updates the nodes successor list by
   quering the successor of all the nodes in
   the list.
*/
func (n *Node) updateSuccList() {
	len := n.getSuccListLen()
	for i := 0; i < len - 1; i++ {
		succ := n.getSuccIndex(i)

		nextSucc, err := n.getSucc(succ)
		if err == nil {
			n.setSuccIndex(nextSucc, (i+1))
		} else {
			if i > 0 {
				n.clearSuccIndex(i)
			}
		}
	}

}

// Setter for successor
func (n *Node) setSuccessor(node shared.NodeInfo) {
	n.update.Lock()

	n.table.fingers[1].node = node
	n.succList[0] = node

	n.update.Unlock()

	//Used for visulization
	//n.updateState()
}


func (n *Node) printSuccList() {
	len := n.getSuccListLen()
	for i := 0; i < len; i++ {
		n.update.RLock()
		n.logger.Info(n.succList[i].Ip)
		n.update.RUnlock()
	}

}

//Getter for successor
func (n *Node) getSuccessor() shared.NodeInfo {
	n.update.RLock()

	succ := n.succList[0]

	if equal := compareAddr(succ, n.succList[0]); !equal {
		n.logger.Error("Successor and successor list not alligned")
	}

	n.update.RUnlock()
	return succ
}

//Sets the successor on the given index in the successor list
func (n *Node) setSuccIndex(node shared.NodeInfo, index int) {
	n.update.Lock()

	n.succList[index] = node

	n.update.Unlock()
}

//Gets the successor on the given index within the successor list
func (n *Node) getSuccIndex(index int) shared.NodeInfo {
	n.update.RLock()

	succ := n.succList[index]

	n.update.RUnlock()

	return succ
}

func (n *Node) getSuccListLen() int {
	n.update.RLock()

	len := len(n.succList)

	n.update.RUnlock()

	return len
}

//Clears the successor on the given index in the successor list
func (n *Node) clearSuccIndex(index int) {
	n.update.Lock()

	n.succList[index].Ip = ""
	n.succList[index].RpcPort = ""
	n.succList[index].HttpPort = ""

	n.update.Unlock()
}

//Getter for prev node
func (n *Node) getPrev() shared.NodeInfo {
	n.prevLock.RLock()

	prev := n.prev

	n.prevLock.RUnlock()

	return prev
}

//Setter for prev node
func (n *Node) setPrev(newPrev shared.NodeInfo) {
	n.prevLock.Lock()

	n.prev = newPrev

	n.prevLock.Unlock()
	//Used for visulization
	//n.updateState()
}

//Compares the address between two nodes
func compareAddr(node1, node2 shared.NodeInfo) bool {
	addr1 := fmt.Sprintf("%s:%s", node1.Ip, node1.HttpPort)
	addr2 := fmt.Sprintf("%s:%s", node2.Ip, node2.HttpPort)

	return addr1 == addr2
}


