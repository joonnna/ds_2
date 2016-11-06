package client

import (
	"sync"
	"runtime"
	"os"
	"io/ioutil"
	"net/http"
	"github.com/joonnna/ds_2/logger"
	"strings"
	"time"
)

type Client struct {
	nodeIp string
	client *http.Client
	log *logger.Logger
	nodes []string
	lock sync.RWMutex
}
var (
	i = 0
)

func (c *Client) getNode() string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	ret := c.nodes[i]
	i = (i + 1) % len(c.nodes)

	return ret
}

func (c *Client) removeNode() string {
	c.lock.Lock()
	defer c.lock.Unlock()

	for i := 0; i < len(c.nodes); i++ {
		ip := c.nodes[i]
		c.nodes = append(c.nodes[:i], c.nodes[i+1:]...)
		return ip
	}
	return ""
}



func (c *Client) addRequest(wg *sync.WaitGroup) {
	ip := c.getNode()

	req, err := http.NewRequest("POST", "http://" + ip + "/addNode", nil)
	if err != nil {
		c.log.Error(err.Error())
	}

	req.Close = true
	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error(err.Error())
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.log.Error(err.Error())
		}
		resp.Body.Close()
		c.lock.Lock()
		c.nodes = append(c.nodes, string(body))
		c.lock.Unlock()
	}
	wg.Done()
}


func (c *Client) shutdownRequest(wg *sync.WaitGroup) {
	ip := c.removeNode()

	req, err := http.NewRequest("POST", "http://" + ip + "/shutdown", nil)
	if err != nil {
		c.log.Error(err.Error())
	}

	req.Close = true
	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error(err.Error())
	} else {
		_, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.log.Error(err.Error())
		}
		resp.Body.Close()
	}
	wg.Done()

}

func (c *Client) neighbourRequest() {
	i := 0
	for {
		time.Sleep(time.Second)

		c.lock.RLock()
		ip := c.nodes[i]
		c.lock.RUnlock()

		req, err := http.NewRequest("GET", "http://" + ip + "/neighbours", nil)
		if err != nil {
			c.log.Error(err.Error())
		}

		req.Close = true
		resp, err := c.client.Do(req)
		if err != nil {
			c.log.Error(err.Error())
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.log.Error(err.Error())
			}
			resp.Body.Close()
			tmp := strings.Split(string(body), "\n")
			c.updateState(ip, tmp[0], tmp[1])
		}

		c.lock.RLock()
		i = ((i + 1) % len(c.nodes))
		c.lock.RUnlock()
	}
}

func Run(nodeIp string) {
	var wg sync.WaitGroup
	numRequests := 100

	runtime.GOMAXPROCS(runtime.NumCPU())

	log := new(logger.Logger)
	log.Init((os.Stdout), "client", 0)

	c := &Client {
		client: &http.Client{},
		nodeIp: nodeIp,
		log: log}

	c.nodes = append(c.nodes, nodeIp)
	c.log.Info("Started client")

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go c.addRequest(&wg)
	}
	wg.Wait()

	//Wait for stabilize
	time.Sleep(time.Second*5)

	for i := 0; i < (numRequests/2); i++ {
		wg.Add(1)
		go c.shutdownRequest(&wg)
	}
	wg.Wait()
}



