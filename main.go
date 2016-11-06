package main

import (
	"github.com/joonnna/ds_2/client"
	"github.com/joonnna/ds_2/node"
	"os"
	"strings"
)


func main() {
	temp := strings.Join(os.Args[1:], "")
	args := strings.Split(temp, ",")

	progType := args[(len(args)-1)]

	switch progType {

		case "node":
			node.Run(args[0])

		case "client":
			client.Run(args[0])
	}
}
