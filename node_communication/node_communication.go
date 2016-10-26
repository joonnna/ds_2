package shared

import(
	"net/rpc"
	"net"
	"strings"
)

type Comm struct {
	Client *rpc.Client
}
/* Inits the rpc server */
func InitRpcServer(address string, api RPC) (net.Listener, error) {
	server := rpc.NewServer()
	err := server.RegisterName("Node", api)
	if err != nil {
		return nil, err
	}

	t, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, err
	}

	l, err := net.ListenTCP("tcp4", t)
	if err != nil {
		return nil, err
	}

	go server.Accept(l)

	return l, nil
}

func setupConn(address string, port string) (*Comm, error) {
	var addr string
	tmp := strings.Split(address, ":")
	if len(tmp) == 0 {
		addr = address + port
	} else {
		addr = tmp[0] + port
	}
	client, err := dialNode(addr)
	if err != nil {
		return nil, err
	}

	return &Comm{Client: client}, nil
}

func dialNode(address string) (*rpc.Client, error) {
	t, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, err
	}

	connection, err := net.DialTCP("tcp4", nil, t)
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
