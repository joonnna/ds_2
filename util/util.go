package util

import (
	"encoding/json"
	"time"
	"math/big"
	"fmt"
	"os"
	"io"
	"crypto/sha1"
	"io/ioutil"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/joonnna/ds_chord/node_communication"
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
	scriptName := "./rocks_list_random_hosts.sh"
	cmd := exec.Command("sh", scriptName, "1")

	nodeAddr, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	return nodeAddr
}


func RpcArgs(key big.Int, value string) shared.Args {
	args := shared.Args {
		Key: key,
		Value: value }

	return args
}


func CreateArgs(nodeAddr string, nodeId big.Int) shared.Args {
	n := shared.NodeInfo{
		Ip: nodeAddr,
		Id: nodeId }

	args := shared.Args{
		Node: n }

	return args
}

