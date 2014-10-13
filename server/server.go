package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	gson "github.com/bitly/go-simplejson"
	bank "github.com/replicasystem/bank"
	"github.com/replicasystem/structs"
)

func SendRequest(server string, request *structs.Request) {
	res1B, err := json.Marshal(request)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+server+"/sync", bytes.NewBuffer(res1B))
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

func queryhandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	if r.Method == "GET" {
		fmt.Fprint(w, "Hello, query")
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Println(body)
		js, _ := gson.NewJson(body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		jaf, _ := js.Get("Requestid").String()
		fmt.Println(res)
		fmt.Println(jaf)
		fmt.Println(b.GetBalance(res))
	}
}

func updatehandler(w http.ResponseWriter, r *http.Request, b *bank.Bank, chain *structs.Chain) {
	fmt.Fprint(w, "Hello, update")

	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Println(body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		fmt.Println(res)
		if res.Transaction == "deposit" {
			res1D := b.Deposit(res)
			SendRequest(chain.Next, res1D)
		}
		if res.Transaction == "withdraw" {
			fmt.Println(b.Withdraw(res))
		}
		//fmt.Println(b.GetBalance(res))
	}

}

func synchandler(w http.ResponseWriter, r *http.Request, b *bank.Bank, chain *structs.Chain) {
	fmt.Fprint(w, "Hello, sync")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		fmt.Println(res)
		b.Set(res)
		if chain.Istail {
			SendRequest("localhost:12345", res)
		} else {
			SendRequest(chain.Next, res)
		}
		//fmt.Println(b.GetBalance())
	}
}

func main() {
	b := bank.Initbank("wellsfargo", "wells")
	chain := structs.Makechain(4, 4001, 3)
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		queryhandler(w, r, b)
	})
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		updatehandler(w, r, b, chain)
	})
	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		synchandler(w, r, b, chain)
	})
	err := http.ListenAndServe(chain.Server, nil)
	if err != nil {
		log.Fatal(err)
	}

}
