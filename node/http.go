package node

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"time"
)

const (
	startPort = 2000
	endPort = 9000
)

func (n *Node) shutDownHandler(w http.ResponseWriter, r *http.Request) {
	n.flagLock.Lock()
	n.shutdownFlag = true
	n.flagLock.Unlock()

	n.wg.Wait()

	n.httpListener.Close()

	//Extra caution if some are still not done
	time.Sleep(time.Second*2)

	n.exitChan <- 1
}

func (n *Node) addHandler(w http.ResponseWriter, r *http.Request) {
	n.flagLock.RLock()
	if n.shutdownFlag {
		n.flagLock.RUnlock()
		w.WriteHeader(http.StatusNotFound)
		return
	}
	n.flagLock.RUnlock()

	n.wg.Add(1)
	addr, err := n.launchNewNode()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		fmt.Fprintf(w, addr)
	}
	n.wg.Done()
}


func (n *Node) neighbourHandler(w http.ResponseWriter, r *http.Request) {
	n.flagLock.RLock()
	if n.shutdownFlag {
		n.flagLock.RUnlock()
		w.WriteHeader(http.StatusNotFound)
		return
	}
	n.flagLock.RUnlock()

	n.wg.Add(1)
	succ := n.getSuccessor()
	prev := n.getPrev()
	addr1 := fmt.Sprintf("%s:%s", succ.Ip, succ.HttpPort)
	addr2 := fmt.Sprintf("%s:%s", prev.Ip, prev.HttpPort)


	retString := fmt.Sprintf("%s\n%s", addr1, addr2)

	fmt.Fprintf(w, retString)
	n.wg.Done()
}

/* Responsible for handling http requests*/
func (n *Node) httpHandler() {
	r := mux.NewRouter()

	r.HandleFunc("/addNode", n.addHandler).Methods("POST")
	r.HandleFunc("/shutdown", n.shutDownHandler).Methods("POST")
	r.HandleFunc("/neighbours", n.neighbourHandler).Methods("GET")

	defer n.httpListener.Close()

	err := http.Serve(n.httpListener, r)
	if err != nil {
		n.logger.Error(err.Error())
		n.logger.Debug("Shutting down http server")
	}
}
