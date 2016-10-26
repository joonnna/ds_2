package node

import (
	"github.com/joonnna/ds_chord/util"
	"github.com/joonnna/ds_chord/chord"
	"github.com/joonnna/ds_chord/logger"
	"runtime"
	"net"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
	"github.com/gorilla/mux"
	"time"
)

func (n *Node) shutDownHandler(w http.ResponseWriter, r *http.Request) {

}

func (n *Node) addHandler(w http.ResponseWriter, r *http.Request) {
	addr = n.launchNewNode()

	fmt.Fprintf(w, addr)
}



func (n Node) neighbourHandler(w http.ResponseWriter, r *http.Request) {

	retString = n.fingerTable.fingers[1].node.Ip + "\n" + n.prev.Ip

	fmt.Fprintf(w, retString)

}

/* Responsible for handling http requests*/
func (n *Node) httpHandler() {
	r := mux.NewRouter()
	r.HandleFunc("/addNode", s.addHandler).Methods("POST")
	r.HandleFunc("/shutdown", s.shutDownHandler).Methods("POST")
	r.HandleFunc("/neighbours", s.neighbourHandler).Methods("GET")

	l, err := net.Listen("tcp4", n.httpPort)
	if err != nil {
		s.log.Error(err.Error())
		os.Exit(1)
	}
	defer l.Close()

	err = http.Serve(l, r)
	if err != nil {
		s.log.Error(err.Error())
		os.Exit(1)
	}
}

/* Inits and runs the chord implementation, responsible for handling requests*/
/*
func Run(nameServer, httpPort, rpcPort string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.DefaultTransport.(*http.Transport).IdleConnTimeout = time.Second * 1
	http.DefaultTransport.(*http.Transport).MaxIdleConns = 10000

	l := new(logger.Logger)
	l.Init((os.Stdout), "Storage", 0)

	storage := &Storage{
		chord: chord.Init(nameServer, httpPort, rpcPort),
		log: l}

	go storage.chord.Run()

	storage.httpHandler(httpPort)
}
*/
