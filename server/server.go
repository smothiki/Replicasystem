package main

import (
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
		js, _ := gson.NewJson(body)
		jaf, _ := js.Get("reqId").String()
		fmt.Println(jaf)
	}
}

func synchandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, update")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Println(body)
		js, _ := gson.NewJson(body)
		jaf, _ := js.Get("reqId").String()
		fmt.Println(jaf)
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
	http.HandleFunc("/sync", synchandler)
	err := http.ListenAndServe(os.Args[1]+":"+os.Args[2], nil)
	if err != nil {
		log.Fatal(err)
	}

}
