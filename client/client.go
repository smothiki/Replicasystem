package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/replicasystem/structs"
)

type ChainList struct {
	head string
	tail string
}

func SendRequest(server string, request *structs.Request) {
	res1B, err := json.Marshal(request)
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
	if r.Method == "POST" {
		fmt.Fprint(w, "Hello, query")
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Println(body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		fmt.Println(res)
	}
}

func main() {
	list := make([]ChainList, 2)
	chain1 := ChainList{
		head: "localhost:4000",
		tail: "localhost:4002",
	}
	list = append(list, chain1)
	res1D := structs.Genrequest(0, "getbalance")
	fmt.Println(res1D)
	//res1D.Account = "f12da044"
	SendRequest(list[0].head, res1D)
	http.HandleFunc("/sync", synchandler)
	err := http.ListenAndServe("localhost:12345", nil)
	if err != nil {
		log.Fatal(err)
	}
}
