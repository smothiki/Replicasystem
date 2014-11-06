package main

import (
	//"bytes"
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"log"
	"net"
	//"net/http"
	"time"

	"github.com/replicasystem/src/commons/structs"
	"github.com/replicasystem/src/commons/utils"
)

const MAXLINE = 1024

type ChainList struct {
	head string
	tail string
}

/* SendRequest sends request (query/update) to server */
func SendRequest(server string, request *structs.Request) {
	res1B, err := json.Marshal(request)
	fmt.Println(string(res1B))

	destIP, destPort := structs.GetIPAndPort(server)
	destAddr := net.UDPAddr{
		Port: destPort,
		IP:   net.ParseIP(destIP),
	}

	conn, err := net.DialUDP("udp", nil, &destAddr)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	defer conn.Close()

	n, err := conn.Write(res1B)
	fmt.Println(n)
}

func createUDPSocket() *net.UDPConn {
	sLocalAddr := utils.Getconfig("client")
	ip, port := structs.GetIPAndPort(sLocalAddr)
	localAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}
	conn, err := net.ListenUDP("udp", &localAddr)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func readResponse(conn *net.UDPConn) *structs.Request {
	buf := make([]byte, MAXLINE)
	n, _, err := conn.ReadFromUDP(buf)

	if err != nil {
		fmt.Println("Error while reading UDP: %s\n", err)
	}

	rqst := &structs.Request{}
	json.Unmarshal(buf[:n], &rqst)

	return rqst
}

// http handler function for listening to sync requests
/*
func synchandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, client")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		fmt.Println("SYNC")
		fmt.Println(res)
		utils.Logoutput("client", res.Requestid, res.Outcome, res.Balance, res.Transaction)
	}
}
*/

// simulates the client and sends request to server

func simulate(chain *ChainList, conn *net.UDPConn) {

	listreqs := structs.GetrequestList(0, "getbalance")
	var dest string
	for _, request := range *listreqs {
		if request.Transaction == "getbalance" {
			dest = chain.tail
		} else {
			dest = chain.head
		}

		// SendRequest(chain.tail, "GET", "query", &request)
		err := utils.Timeout("timeout", time.Duration(5)*time.Second, func() {
			SendRequest(dest, &request)
			fmt.Println(readResponse(conn))
		})
		if err != nil {
			fmt.Println("timeout")
		}
	}
}

func main() {
	chain1 := &ChainList{
		head: "localhost:4001",
		tail: "localhost:4003",
	}
	fmt.Println("start server")
	conn := createUDPSocket()

	//go simulate(chain1, conn)
	simulate(chain1, conn)
	//re := structs.Request{"1.1.1", "12", 5, "deposit", ""}
	//SendRequest("127.0.0.1:4001", &re)
	//readResponse(conn)
	//http.HandleFunc("/sync", synchandler)
	/*err := http.ListenAndServe(utils.Getconfig("client"), nil)
	if err != nil {
		log.Fatal(err)
	}*/
}
