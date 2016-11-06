package node

import (
	"net"
	"strings"
	"os"
	"github.com/joonnna/ds_2/node_communication"
	"github.com/joonnna/ds_2/util"
	"github.com/joonnna/ds_2/logger"
	"sync"
	"math/big"
	"net/http"
	"os/exec"
	"fmt"
	"runtime"
	"strconv"
	"time"
	"errors"
)

/* Node defenition */
type Node struct {
	//Node information
	info shared.NodeInfo

	//Flag used when shutdown request is received
	shutdownFlag bool
	flagLock sync.RWMutex

	//Channel receiving the httpPort the bootstrapped node is listening on
	queue chan shared.BootStrapNode

	//Successor list, periodically updated
	succList []shared.NodeInfo

	//Node logger
	logger *logger.Logger

	//RPC listener
	listener net.Listener

	httpListener net.Listener

	prev shared.NodeInfo

	//Lock for prev node info
	prevLock sync.RWMutex

	table fingerTable

	//Waitgroup used for shutting down http server
	wg *sync.WaitGroup

	//Chan used during shutdown of node
	exitChan chan int

	//Lock used for successor list and successor
	update sync.RWMutex
}

var (
	ErrBootNode = errors.New("Didn't receive bootstrap node info")
)

/* Inits and returns the node object
   nameserver: ip address of the nameserver
   httpPort: Port for http communication
   rpcPort: Port for rpc communication

   Returns node object
*/

func initNode() *Node {
	hostName, _ := os.Hostname()
	hostName = strings.Split(hostName, ".")[0]
	http.DefaultTransport.(*http.Transport).MaxIdleConns = 1000

	log := new(logger.Logger)
	log.Init((os.Stdout), hostName, 0)

	info := shared.NodeInfo {
		Ip: hostName }

	n := &Node {
		succList: make([]shared.NodeInfo, 10),
		info: info,
		queue: make(chan shared.BootStrapNode),
		exitChan: make(chan int),
		logger: log,
		wg: &sync.WaitGroup{}}


	l, port, err := shared.InitRpcServer(hostName, n)
	if err != nil {
		n.logger.Error(err.Error())
		os.Exit(1)
	}

	n.info.RpcPort = strconv.Itoa(port)
	n.listener = l

	idStr := fmt.Sprintf("%s:%s", hostName, n.info.RpcPort)
	tmp := util.ConvertKey(idStr)
	id := new(big.Int)
	id.SetBytes(tmp)

	n.info.Id = *id

	n.initFingerTable()
	return n
}

/* Launches a new node on a random node in the cluster
   Returns the ip:port pair of the newly started node.
*/
func (n *Node) launchNewNode() (string, error)  {
	nodeAddr := util.GetNodeAddr()

	path := "./go/src/github.com/joonnna/ds_2/main.go"

	addr := fmt.Sprintf("%s:%s", n.info.Ip, n.info.RpcPort)

	command := fmt.Sprintf("go run %s %s,node", path, addr)

	cmd := exec.Command("ssh", "-T", nodeAddr, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()

	address, err := n.getBootStrapNode()
	if err != nil {
		n.logger.Error(err.Error())
		return "", err
	}

	return fmt.Sprintf("%s:%s", address.Ip, address.HttpPort), nil
}


/* Waits for the bootstrapped node to send its address
   After a timeout it returns an error
*/
func (n *Node) getBootStrapNode() (shared.BootStrapNode, error) {
	select {
		case address := <- n.queue:
			return address, nil
		case <- time.After(time.Second*5):
			return shared.BootStrapNode{}, ErrBootNode
	}
}


/* Boostrap nodes sends it address to its creation node
   through a RPC call.
*/
func (n *Node) sendBootStrapNode(address string) {
	tmp := strings.Split(address, ":")

	ip := tmp[0]
	port := tmp[1]

	args := shared.BootStrapNode {
		Ip: n.info.Ip,
		HttpPort: n.info.HttpPort }
	reply := &shared.Reply{}

	err := shared.SingleCall("Node.AddBootStrapNode", ip, port, args, reply)
	if err != nil {
		n.logger.Error(err.Error())
	}
}

/* Joins the network and calls stabilize and fix fingers periodically.
	node: the node to run
*/
func Run(nodeip string) {
	n := initNode()
	defer n.listener.Close()

	runtime.GOMAXPROCS(runtime.NumCPU())
	http.DefaultTransport.(*http.Transport).IdleConnTimeout = time.Second * 1
	http.DefaultTransport.(*http.Transport).MaxIdleConns = 10000

	l, port, err := util.FindPort(startPort, endPort)
	if err != nil {
		n.logger.Error(err.Error())
		os.Exit(1)
	}
	n.httpListener = l
	n.info.HttpPort = strconv.Itoa(port)

	n.join(nodeip)

	//First node in network case
	if nodeip != "" {
		n.sendBootStrapNode(nodeip)
	}


	go n.httpHandler()

	//Used for visualization
	//n.add()
	//defer n.remove()

	go n.stabilize()

	n.logger.Info(fmt.Sprintf("Started node, http: %s rpc: %s", n.info.HttpPort, n.info.RpcPort))
	<-n.exitChan
	n.logger.Info("Shutting down")
}
