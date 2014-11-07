package main

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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
func SendRequest(server string, request *structs.Request, port int) {
	destIP, destPort := structs.GetIPAndPort(server)
	destAddr := net.UDPAddr{
		Port: destPort,
		IP:   net.ParseIP(destIP),
	}

	localAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &destAddr)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	defer conn.Close()

	request.Client = localAddr
	request.Time = time.Now().String()
	res1B, err := json.Marshal(request)
	fmt.Println("sendReq", string(res1B))

	n, err := conn.Write(res1B)
	fmt.Println(n)
}

func createUDPSocket(client string) *net.UDPConn {
	ip, port := structs.GetIPAndPort(client)
	fmt.Println("createUDP", client)
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
		fmt.Println("Error while reading UDP", err)
	}

	rqst := &structs.Request{}
	json.Unmarshal(buf[:n], &rqst)
	go utils.Logoutput("client", rqst.Requestid, rqst.Outcome, rqst.Balance, rqst.Transaction)

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

func alterChainHandler(w http.ResponseWriter, r *http.Request, chain *structs.Chain) {
	fmt.Fprintf(w, "changed")
	body, _ := ioutil.ReadAll(r.Body)
	newHeadTail := &structs.ClientNotify{}
	json.Unmarshal(body, &newHeadTail)
	if newHeadTail.Head != "" {
		chain.Head = newHeadTail.Head
		fmt.Println("newHead", chain.Head)
	} else if newHeadTail.Tail != "" {
		chain.Tail = newHeadTail.Tail
		fmt.Println("newTail", chain.Tail)
	}
}

// simulates the client and sends request to server
func simulate(chain *structs.Chain, conn *net.UDPConn, port int) {
	listreqs := structs.GetrequestList(0, "getbalance")
	var dest string
	for _, request := range *listreqs {
		if request.Transaction == "getbalance" {
			dest = chain.Tail
		} else {
			dest = chain.Head
		}
		fmt.Println("dest", dest)

		// SendRequest(chain.tail, "GET", "query", &request)
		err := utils.Timeout("timeout", time.Duration(5)*time.Second, func() {
			SendRequest(dest, &request, port)
			fmt.Println("mYR", readResponse(conn))
		})
		if err != nil {
			fmt.Println("timeout")
		}
	}
}

func main() {
	port, _ := strconv.Atoi(os.Args[1])
	utils.SetConfigFile("config.json")
	series, _ := strconv.Atoi(utils.Getconfig("chian1series"))
	lenservers, _ := strconv.Atoi(utils.Getconfig("chainlength"))
	curseries := int(port / 1000)
	series = series + (curseries - series)
	chain := structs.Makechain(series, port, lenservers)
	fmt.Println("start client")
	conn := createUDPSocket("127.0.0.1:" + os.Args[1])
	//go simulate(chain, conn)
	go simulate(chain, conn, port)
	//re := structs.Request{"1.1.1", "12", 5, "deposit", ""}
	//SendRequest("127.0.0.1:4001", &re)
	//readResponse(conn)
	http.HandleFunc("/alterChain", func(w http.ResponseWriter, r *http.Request) {
		alterChainHandler(w, r, chain)
	})
	err := http.ListenAndServe("127.0.0.1:"+os.Args[1], nil)
	if err != nil {
		log.Fatal(err)
	}
}
