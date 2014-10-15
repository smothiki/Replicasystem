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
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("1" + string(body))
}

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
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func synchandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, client")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		utils.Logoutput("client", res.Requestid, res.Outcome, res.Balance)
	}
}

func simulate(chain *ChainList) {

	listreqs := structs.GetTestreqs()
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
	err := http.ListenAndServe("localhost:10001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
