package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/replicasystem/src/commons/structs"
	"github.com/replicasystem/src/commons/utils"
)

type ChainList struct {
	head string
	tail string
}

// sends update request to server

func SendUpdate(server string, request *structs.Request) {
	res1B, err := json.Marshal(request)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://"+server+"/update", bytes.NewBuffer(res1B))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error : %s", err)
	}
	_, _ = ioutil.ReadAll(resp.Body)
}

// send query request to tail
func Sendquery(server string, request *structs.Request) {
	res1B, err := json.Marshal(request)
	fmt.Println(string(res1B))
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+server+"/query", bytes.NewBuffer(res1B))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error : %s", err)
	}
	_, _ = ioutil.ReadAll(resp.Body)
}

// http handler function for listening to sync requests

func synchandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, client")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		utils.Logoutput("client", res.Requestid, res.Outcome, res.Balance)
	}
}

// simulates the client and sends request to server

func simulate(chain *ChainList) {

	listreqs := structs.GetrequestList(0, "getbalance")
	for _, request := range *listreqs {
		if request.Transaction == "getbalance" {
			// SendRequest(chain.tail, "GET", "query", &request)
			err := utils.Timeout("timeout", time.Duration(5)*time.Second, func() { Sendquery(chain.tail, &request) })
			if err != nil {
				fmt.Println("timeout")
			}
		} else {
			// SendUpdate(chain.head, &request)
			err := utils.Timeout("timeout", time.Duration(5)*time.Second, func() { SendUpdate(chain.head, &request) })
			if err != nil {
				fmt.Println("timeout")
			}
		}
	}
}

func main() {
	chain1 := &ChainList{
		head: "localhost:4001",
		tail: "localhost:4003",
	}
	go simulate(chain1)
	fmt.Println("start servver")
	http.HandleFunc("/sync", synchandler)
	err := http.ListenAndServe(utils.Getconfig("client"), nil)
	if err != nil {
		log.Fatal(err)
	}
}
