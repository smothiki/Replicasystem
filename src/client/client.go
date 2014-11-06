package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/replicasystem/src/commons/structs"
	"github.com/replicasystem/src/commons/utils"
)

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
		utils.Logoutput("client", res.Requestid, res.Outcome, res.Balance, res.Transaction)
	}
}

// simulates the client and sends request to server

func simulate(chain *structs.Chain) {

	listreqs := structs.GetrequestList(0, "getbalance")
	for _, request := range *listreqs {
		if request.Transaction == "getbalance" {
			// SendRequest(chain.tail, "GET", "query", &request)
			err := utils.Timeout("timeout", time.Duration(5)*time.Second, func() { Sendquery(chain.Tail, &request) })
			if err != nil {
				fmt.Println("timeout")
			}
		} else {
			// SendUpdate(chain.head, &request)
			err := utils.Timeout("timeout", time.Duration(5)*time.Second, func() { SendUpdate(chain.Head, &request) })
			if err != nil {
				fmt.Println("timeout")
			}
		}
	}
}

func main() {
	port, _ := strconv.Atoi(os.Args[1])
	series, _ := strconv.Atoi(utils.Getconfig("chian1series"))
	lenservers, _ := strconv.Atoi(utils.Getconfig("chainlength"))
	curseries := int(port / 1000)
	series = series + (curseries - series)
	chain := structs.Makechain(series, port, lenservers)
	go simulate(chain)
	fmt.Println("start servver")
	http.HandleFunc("/sync", synchandler)
	err := http.ListenAndServe(chain.Client, nil)
	if err != nil {
		log.Fatal(err)
	}
}
