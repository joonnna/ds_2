package main

import (
	"os/exec"
	"log"
	"os"
	"syscall"
	"os/signal"
	"strings"
	"strconv"
	"time"
	"fmt"
)
/* Ports to use */
var (
	http = 2345
	rpc = 7453
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
func launch(nodeName, path, nodeIp)  {
	var command string

	httpPort := ":" + strconv.Itoa(http)
	rpcPort := ":" + strconv.Itoa(rpc)

	command = "go run " + path + " " + httpPort + "," + rpcPort + "," + " "

	cmd := exec.Command("ssh", "-T", nodeName, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
}


func main () {
	nodeAddr := GetNodeAddr()

	path := "./go/src/github.com/joonnna/ds_chord/main.go"

	launch(nodeAddr, path)

	/* Wait for CTRL-C then shut all nodes down*/
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	cleanUp()
}
