package node

import (
	"github.com/joonnna/ds_2/node_communication"
	"github.com/joonnna/ds_2/util"
	"errors"
	"math/big"
)

var (
	ErrGet = errors.New("Unable to get key")
	ErrPut = errors.New("Unable to put key")
	ErrFound = errors.New("Unable to find entry in finger table")
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

	curr_succ := n.getSuccessor()

	if equal := compareAddr(curr_succ, n.info); equal {
		reply.Next = n.info
		reply.Prev = n.getPrev()
		return nil
	}
	node, err := n.findPreDecessor(test.Id)
	if err != nil {
		n.logger.Error(err.Error())
	}

	succ, err := n.getSucc(node)
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
func (n *Node) ClosestPrecedingFinger(args shared.Search, reply *shared.Reply) error {
	cmpId := new(big.Int)
	cmpId.SetBytes(args.Id)
	for i := (lenOfId-1); i >= 1; i-- {
		entry := n.table.fingers[i].node.Id
		if entry.BitLen() != 0 && util.BetweenNodes(n.info.Id, *cmpId, entry) {
			reply.Next = n.table.fingers[i].node
			return nil
		}
	}

	reply.Next = n.info

	return ErrFound
}

/* Populates the reply with the nodes predecessor */
func (n *Node) GetPreDecessor(args int, reply *shared.Search) error {
	prev := n.getPrev()

	reply.Id = prev.Id.Bytes()
	reply.Ip = prev.Ip
	reply.RpcPort = prev.RpcPort
	reply.HttpPort = prev.HttpPort

	return nil
}

/* Populates the reply with the nodes successor */
func (n *Node) GetSuccessor(args int, reply *shared.Search) error {
	succ := n.getSuccessor()

	reply.Id = succ.Id.Bytes()
	reply.Ip = succ.Ip
	reply.RpcPort = succ.RpcPort
	reply.HttpPort = succ.HttpPort

	return nil
}
/* Checks if the given node id is the new predecessor */
func (n *Node) Notify(args shared.Search, reply *shared.Reply) error {
	tmp := new(big.Int)
	tmp.SetBytes(args.Id)

	node := shared.NodeInfo {
		Id: *tmp,
		Ip: args.Ip,
		RpcPort: args.RpcPort,
		HttpPort: args.HttpPort }

	prev := n.getPrev()
	if dead := util.PingNode(prev.Ip, prev.RpcPort); dead {
		n.setPrev(node)
	}

	equal := compareAddr(prev, n.info)
	if equal || util.BetweenNodes(prev.Id, n.info.Id, *tmp) {
		n.setPrev(node)
	}

	curr_succ := n.getSuccessor()
	if equal := compareAddr(curr_succ, n.info); equal {
		n.setSuccessor(node)
	}

	return nil
}

/* Rpc call to send the bootstrap node address */
func (n *Node) AddBootStrapNode(args shared.BootStrapNode, reply *shared.Reply) error {
	n.queue <- args

	return nil
}
