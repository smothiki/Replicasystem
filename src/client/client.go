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

func SendRequest(server, method, api string, request *structs.Request) {
	res1B, _ := json.Marshal(request)
	client := &http.Client{}
	req, _ := http.NewRequest(method, "http://"+server+"/"+api, bytes.NewBuffer(res1B))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	if api == "query" {
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error : %s", err)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		utils.Logoutput("client", res.Requestid, res.Outcome, res.Balance)
	} else {
		go func() {
			_, err := client.Do(req)
			if err != nil {
				fmt.Printf("Error : %s", err)
			}
		}()
	}

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

	listreqs := structs.GetrequestList(3, "getbalance")
	for _, request := range *listreqs {
		if request.Transaction == "getbalance" {
			// SendRequest(chain.tail, "GET", "query", &request)
			err := utils.Timeout("timeout", time.Duration(5)*time.Second, func() { SendRequest(chain.tail, "GET", "query", &request) })
			if err != nil {
				fmt.Println("timeout")
			}
		} else {
			//SendRequest(chain.head, "POST", "update", &request)
			err := utils.Timeout("timeout", time.Duration(5)*time.Second, func() { SendRequest(chain.head, "POST", "update", &request) })
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

	//res1D.Account = "f12da044"
	go simulate(chain1)
	fmt.Println("start servver")
	http.HandleFunc("/sync", synchandler)
	err := http.ListenAndServe("localhost:10001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
