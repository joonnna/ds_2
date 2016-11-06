package util

import (
	"encoding/json"
	"time"
	"math/big"
	"fmt"
	"os"
	"io"
	"strings"
	"crypto/sha1"
	"io/ioutil"
	"github.com/gorilla/mux"
	"net/http"
	"net"
//	"github.com/joonnna/ds_2/node_communication"
	"os/exec"
)


/* With  Upper Inlcude */
func InKeySpace(start, end, newId big.Int) bool {
	startEndCmp := start.Cmp(&end)

	if startEndCmp == -1 {
		if start.Cmp(&newId) == -1 && end.Cmp(&newId) >= 0 {
			return true
		} else {
			return false
		}
	} else {
		if start.Cmp(&newId) == -1 || end.Cmp(&newId) >=0 {
			return true
		} else {
			return false
		}
	}
}

/* Without include */
func BetweenNodes(start, end, newId big.Int) bool {
	startEndCmp := start.Cmp(&end)

	if startEndCmp == -1 {
		if start.Cmp(&newId) == -1 && end.Cmp(&newId) == 1 {
			return true
		} else {
			return false
		}
	} else {
		if start.Cmp(&newId) == -1 || end.Cmp(&newId) == 1 {
			return true
		} else {
			return false
		}
	}
}

func ConvertKey(key string) []byte {
	h := sha1.New()
	io.WriteString(h, key)

	return h.Sum(nil)
}

func ConvertToBigInt(bytes []byte) big.Int {
	ret := new(big.Int)
	ret.SetBytes(bytes)
	return *ret
}

func GetNode(curNode string, nameServer string) (string, error) {
	list, err := GetNodeList(nameServer)
	if err != nil {
		return "", err
	}
	for _, ip := range list {
		if ip != curNode {
			return ip, nil
		}
	}
	return "", nil
}


func GetKey(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["key"]
}

func CheckInterrupt() {
	for {
		msg, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println(err.Error())
		}

		if string(msg) == "kill" {
			os.Exit(1)
		}
	}
}

func GetNodeList(nameServer string) ([]string, error)  {
	var nodeIps []string

	timeout := time.Duration(5 * time.Second)
	client := &http.Client{Timeout : timeout}

	r, err := client.Get(nameServer)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &nodeIps)
	if err != nil {
		return nil, err
	}
	return nodeIps, nil
}

func GetNodeAddr() string {
	scriptName := "/home/jmi021/go/src/github.com/joonnna/ds_2/launch/rocks_list_random_hosts.sh"
	cmd := exec.Command("sh", scriptName, "1")

	nodeAddr, err := cmd.Output()
	if err != nil {
		return "Couldn't retrieve node address"
	}

	ret := strings.Split(string(nodeAddr), " \n")

	return ret[0]
}


func FindPort(startPort, endPort int) (net.Listener, int, error) {

	port := startPort

	for {
		listenAddr := fmt.Sprintf(":%d", port)
		l, err := net.Listen("tcp4", listenAddr)
		if err == nil {
			return l, port, nil
		}

		if port > endPort {
			return nil, 0, err
		}
		port += 1
	}
}

func PingNode(ip, port string) bool {
	timeout := time.Second * 3
	addr := fmt.Sprintf("%s:%s", ip, port)
	conn, err := net.DialTimeout("tcp4", addr, timeout)
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

