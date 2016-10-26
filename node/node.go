package node

import (
	"net"
	"strings"
	"os"
	"github.com/joonnna/ds_chord/node_communication"
	"github.com/joonnna/ds_chord/util"
	"github.com/joonnna/ds_chord/logger"
	"sync"
	"math/big"
	"net/http"
	"time"
)
/* Node defenition */
type Node struct {
	data map[string]string
	Ip string
	id big.Int
	NameServer string
	RpcPort string
	httpPort string
	logger *logger.Logger

	listener net.Listener

	prev shared.NodeInfo

	table fingerTable

	update sync.RWMutex
}
/* Inits and returns the node object
   nameserver: Ip address of the nameserver
   httpPort: Port for http communication
   rpcPort: Port for rpc communication

   Returns node object
*/
func InitNode(httpPort, rpcPort string) *Node {
	hostName, _ := os.Hostname()
	hostName = strings.Split(hostName, ".")[0]
	http.DefaultTransport.(*http.Transport).MaxIdleConns = 1000
	tmp := util.ConvertKey(hostName)

	log := new(logger.Logger)
	log.Init((os.Stdout), hostName, 0)

	id := new(big.Int)
	id.SetBytes(tmp)

	n := &Node {
		id: *id,
		Ip: hostName,
		logger: log,
		httpPort: httpPort,
		RpcPort: rpcPort,
		data: make(map[string]string) }


	l, err := shared.InitRpcServer(hostName + rpcPort, n)
	if err != nil {
		n.logger.Error(err.Error())
		os.Exit(1)
	}

	n.listener = l

	n.initFingerTable()
	return n
}

func (n *Node) launchNewNode() {
	var command string

	nodeAddr = GetNodeAddr()

	path := "./go/src/github.com/joonnna/ds_chord/main.go"

	command = "go run " + path + " " + n.httpPort + "," + n.rpcPort + "," + n.Ip

	cmd := exec.Command("ssh", "-T", nodeName, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()

	return (nodeAddr + n.httpPort)
}

/* Joins the network and calls stabilize and fix fingers periodically.
	node: the node to run
*/
func Run(httpPort, rpcPort, nodeIp string) {
	n = initNode(httpPort, rpcPort)

	defer n.listener.Close()

	runtime.GOMAXPROCS(runtime.NumCPU())
	http.DefaultTransport.(*http.Transport).IdleConnTimeout = time.Second * 1
	http.DefaultTransport.(*http.Transport).MaxIdleConns = 10000

	go n.httpHandler()

	n.join(nodeIp)

//	n.add()
//	go n.updateState()

	go n.fixFingers()
	for {
		n.stabilize()
		time.Sleep(time.Second * 1)
	}
}
