package node

import (
	"github.com/joonnna/ds_chord/node_communication"
	"github.com/joonnna/ds_chord/util"
	"errors"
	"math/big"
)

var (
	ErrGet = errors.New("Unable to get key")
	ErrPut = errors.New("Unable to put key")

)

/* Rpc function to find the successor of the given id.
   Also finds the predecessor of the given id.
   args: contains the search id
   reply: populated with successor and predecessor information
*/
func (n *Node) FindSuccessor(args shared.Search, reply *shared.Reply) error {
	tmp := new(big.Int)
	tmp.SetBytes(args.Id)
	test := shared.NodeInfo {
		Ip: args.Ip,
		Id: *tmp }

	if n.table.fingers[1].node.Ip == n.Ip {
		reply.Next.Ip = n.Ip
		reply.Next.Id = n.id
		reply.Prev = n.prev
		return nil
	}
	node, err := n.findPreDecessor(test.Id)
	if err != nil {
		n.logger.Error(err.Error())
		return err
	}

	succ, err := n.getSucc(node.Ip)
	if err != nil {
		return err
	}

	reply.Next = succ
	reply.Prev = node

	return nil
}


/* Rpc function to find the closest preceding finger in the fingertable for the given id.
   Finds node in the fingertable entry which is in the keyspace between itself and the given id.
   reply: populated with the closest preceding finger.
   args: contains the search id
*/
func (n *Node) ClosestPrecedingFinger(args shared.Search, reply *shared.Reply) error{
	cmpId := new(big.Int)
	cmpId.SetBytes(args.Id)
	for i := (lenOfId-1); i >= 1; i-- {
		entry := n.table.fingers[i].node.Id
		if entry.BitLen() != 0 && util.BetweenNodes(n.id, *cmpId, entry) {
			reply.Next = n.table.fingers[i].node
			return nil
		}
	}

	reply.Next.Id = n.id
	reply.Next.Ip = n.Ip

	return nil
}

/* Populates the reply with the nodes predecessor */
func (n *Node) GetPreDecessor(args int, reply *shared.Search) error {
	id := n.prev.Id.Bytes()
	reply.Id = id
	reply.Ip = n.prev.Ip
	return nil
}

/* Populates the reply with the nodes successor */
func (n *Node) GetSuccessor(args int, reply *shared.Search) error {
	id := n.table.fingers[1].node.Id.Bytes()
	reply.Id = id
	reply.Ip = n.table.fingers[1].node.Ip
	return nil
}
/* Checks if the given node id is the new predecessor */
func (n *Node) Notify(args shared.Search, reply *shared.Reply) error {
	tmp := new(big.Int)
	tmp.SetBytes(args.Id)

	node := shared.NodeInfo {
		Id: *tmp,
		Ip: args.Ip	}

	if n.prev.Ip == n.Ip || util.BetweenNodes(n.prev.Id, n.id, *tmp) {
		n.prev = node
	}

	if n.table.fingers[1].node.Ip == n.Ip {
		n.table.fingers[1].node = node
	}

	return nil
}
