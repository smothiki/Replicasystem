package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/replicasystem/structs"
)

func main() {
	res1D := structs.Genrequest(0, "getbalance")
	fmt.Println(res1D)
	res1D.Account = "f12da044"
	res1B, err := json.Marshal(res1D)
	fmt.Println(err)
	fmt.Println(string(res1B))
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:4001/query", bytes.NewBuffer(res1B))
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
