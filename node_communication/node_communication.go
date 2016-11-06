package shared

import(
	"github.com/joonnna/ds_2/util"
	"net/rpc"
	"net"
	"time"
	"fmt"
)

type Comm struct {
	Client *rpc.Client
}

const (
	startPort = 1025
	endPort = 9000
)

/* Inits the rpc server */
func InitRpcServer(address string, api RPC) (net.Listener, int, error) {
	server := rpc.NewServer()
	err := server.RegisterName("Node", api)
	if err != nil {
		return nil, 0, err
	}

	l, port, err := util.FindPort(startPort, endPort)
	if err != nil {
		return l, 0, err
	}

	go server.Accept(l)

	return l, port, nil
}

func setupConn(address string, port string) (*Comm, error) {
	addr := fmt.Sprintf("%s:%s", address, port)

	client, err := dialNode(addr)
	if err != nil {
		return nil, err
	}

	return &Comm{Client: client}, nil
}

func dialNode(address string) (*rpc.Client, error) {
	connection, err := net.DialTimeout("tcp4", address, time.Second)
	if err != nil {
		return nil, err
	}

	return rpc.NewClient(connection), nil
}
/* Wrapper for all rpc calls
   method: rpc method to execute
   address: address of the node to execute the method on
   args: arguments to the method
   reply: return values of the method to be executed
*/
func SingleCall(method string, address string, port string, args interface{}, reply interface{}) error {
	c, err := setupConn(address, port)
	if err != nil {
		return err
	}

	err = c.Client.Call(method, args, reply)
	if err != nil {
		return err
	}

	err = c.Client.Close()
	if err != nil {
		return err
	}

	return nil
}
