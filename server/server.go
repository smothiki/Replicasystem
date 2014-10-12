package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	gson "github.com/bitly/go-simplejson"
	bank "github.com/replicasystem/bank"
	"github.com/replicasystem/structs"
)

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

func updatehandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprint(w, "Hello, update")

	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Println(body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		fmt.Println(res)
		if res.Transaction == "deposit" {
			res1D := b.Deposit(res)
			res1B, err := json.Marshal(res1D)
			fmt.Println(res1B)
			client := &http.Client{}
			req, _ := http.NewRequest("POST", "http://localhost:4001/sync", bytes.NewBuffer(res1B))
			req.Header = http.Header{
				"accept": {"application/json"},
			}
			_, err = client.Do(req)
			if err != nil {
				fmt.Printf("Error : %s", err)
			}
		}
		if res.Transaction == "withdraw" {
			fmt.Println(b.Withdraw(res))
		}
		//fmt.Println(b.GetBalance(res))
	}

}

func synchandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprint(w, "Hello, sync")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Reply{}
		json.Unmarshal(body, &res)
		fmt.Println(res)
		b.Set(res)
		//fmt.Println(b.GetBalance())
	}
}

func main() {
	b := bank.Initbank("wellsfargo", "wells")
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		queryhandler(w, r, b)
	})
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		updatehandler(w, r, b)
	})
	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		synchandler(w, r, b)
	})
	err := http.ListenAndServe(os.Args[1]+":"+os.Args[2], nil)
	if err != nil {
		log.Fatal(err)
	}

}
