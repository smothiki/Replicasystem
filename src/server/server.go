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

	bank "github.com/replicasystem/src/commons/bank"
	"github.com/replicasystem/src/commons/structs"
	"github.com/replicasystem/src/commons/utils"
)

func SendRequest(server string, request *structs.Request) {
	res1B, err := json.Marshal(request)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://"+server+"/sync", bytes.NewBuffer(res1B))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	_, err = client.Do(req)
	if err != nil {
		fmt.Printf("Error : %s", err)
	}
}

func queryhandler(w http.ResponseWriter, r *http.Request, b *bank.Bank, chain *structs.Chain) {
	if r.Method == "GET" {
		fmt.Fprint(w, "Hello, query")
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		res1 := b.GetBalance(res)
		utils.Logoutput("tail", res1.Requestid, res1.Outcome, res1.Balance, res1.Transaction)
		SendRequest(chain.Client, res1)
	}
}

func updatehandler(w http.ResponseWriter, r *http.Request, b *bank.Bank, chain *structs.Chain) {
	fmt.Fprint(w, "Hello, update")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		if res.Transaction == "deposit" {
			res1D := b.Deposit(res)
			fmt.Println("inside deposit" + chain.Next)
			SendRequest(chain.Next, res1D)
			utils.Logoutput(chain.Server, res1D.Requestid, res1D.Outcome, res1D.Balance, res1D.Transaction)
		}
		if res.Transaction == "withdraw" {
			res1D := b.Withdraw(res)
			fmt.Println("inside deposit" + chain.Next)
			SendRequest(chain.Next, res1D)
			utils.Logoutput(chain.Server, res1D.Requestid, res1D.Outcome, res1D.Balance, res1D.Transaction)
		}
	}
}

func synchandler(w http.ResponseWriter, r *http.Request, b *bank.Bank, chain *structs.Chain) {
	fmt.Fprint(w, "Hello, sync")
	fmt.Println("hello syncss")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		fmt.Println(res)
		b.Set(res)
		utils.Logoutput(chain.Server, res.Requestid, res.Outcome, res.Balance, res.Transaction)
		if chain.Istail {
			fmt.Println("inside clientsent" + chain.Next)
			time.Sleep(6000 * time.Millisecond)
			SendRequest(chain.Client, res)
		} else {
			fmt.Println("inside deposit" + chain.Next)
			SendRequest(chain.Next, res)
		}
	}
}

func main() {
	b := bank.Initbank("wellsfargo", "wells")
	port, _ := strconv.Atoi(os.Args[1])
	series, _ := strconv.Atoi(utils.Getconfig("chian1series"))
	lenservers, _ := strconv.Atoi(utils.Getconfig("chainlength"))
	curseries := int(port / 1000)
	series = series + (curseries - series)
	chain := structs.Makechain(series, port, lenservers)
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		queryhandler(w, r, b, chain)
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
