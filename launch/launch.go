package main

import (
	"github.com/joonnna/ds_2/util"
	"os/exec"
	"log"
	"os"
	"syscall"
	"os/signal"
	"strings"
	"time"
	"fmt"
	"runtime"
)

func cleanUp() {
	cmd := exec.Command("sh", "/share/apps/bin/cleanup.sh")
	cmd.Run()
}

func sshToNode(ip string) {
	cmd := exec.Command("ssh", ip)

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
/* Launches the given application(client, nameserver or node)
   nodeName: address of the node to launch
   path: path to the executable to launch
   nameserver: address of the nameserver, empty if application is the nameserver
   flag: -1 if launching the client
   */
func launchClient(nodeName, path, nodeIp string)  {
	command := fmt.Sprintf("go run %s %s,client", path, nodeIp)

	cmd := exec.Command("ssh", "-T", nodeName, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
}


func launchNode(nodeName, path, nodeIp string)  {
	command := fmt.Sprintf("go run %s %s,node", path, nodeIp)

	cmd := exec.Command("ssh", "-T", nodeName, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
}


func main () {
	runtime.GOMAXPROCS(runtime.NumCPU())

	hostName, _ := os.Hostname()
	hostName = strings.Split(hostName, ".")[0]

	nodeAddr := util.GetNodeAddr()
	clientAddr := util.GetNodeAddr()

	//startup := make(chan string)

	path := "./go/src/github.com/joonnna/ds_2/main.go"

	launchNode(nodeAddr, path, "")
	//go node.Run("", startup)
	//httpPort := <- startup
	//fmt.Println("Received port")

	time.Sleep(time.Second * 3)

	address := fmt.Sprintf("%s:%s", nodeAddr, "2000")

	launchClient(clientAddr, path, address)

	/* Wait for CTRL-C then shut all nodes down*/
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	cleanUp()
}
