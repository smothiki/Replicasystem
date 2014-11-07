package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	bank "github.com/replicasystem/src/commons/bank"
	"github.com/replicasystem/src/commons/structs"
	"github.com/replicasystem/src/commons/utils"
)

const MAXLINE = 1024
const SENT_HEALTH_INTERVAL = 1000

func connectToMaster(port int) *net.UDPConn {
	masterAddr := utils.Getconfig("master")
	localAddr := net.UDPAddr{
		Port: port + 100,
		IP:   net.ParseIP("127.0.0.1"),
	}
	destIP, destPort := structs.GetIPAndPort(masterAddr)
	destAddr := net.UDPAddr{
		Port: destPort,
		IP:   net.ParseIP(destIP),
	}
	conn, err := net.DialUDP("udp", &localAddr, &destAddr)

	if err != nil {
		fmt.Printf("ERROR while connecting to master: %s\n", err)
	}
	return conn
}

func sendOnlineMsg(conn *net.UDPConn) {
	for {
		msg, _ := json.Marshal(1)
		_, err := conn.Write(msg)
		if err != nil {
			fmt.Println("ERROR while sending online msg to master: %s\n", err)
		}
		time.Sleep(SENT_HEALTH_INTERVAL * time.Millisecond)
	}
}

/* SendRequest send request to successor */
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

/* SendReply sends reply to client */
func SendReply(client string, request *structs.Request, port int) {
	res1B, err := json.Marshal(request)
	destIP, destPort := structs.GetIPAndPort(client)
	/*localAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}*/

	destAddr := net.UDPAddr{
		Port: destPort,
		IP:   net.ParseIP(destIP),
	}

	conn, err := net.DialUDP("udp", nil, &destAddr)

	if err != nil {
		fmt.Printf("ERROR while connecting to client: %s\n", err)
	}

	defer conn.Close()

	_, err = conn.Write(res1B)
	if err != nil {
		fmt.Println("ERROR while sending reply to client: %s\n", err)
	}
}

func synchandler(w http.ResponseWriter, r *http.Request, b *bank.Bank, chain *structs.Chain, port int) {
	fmt.Fprint(w, "Hello, sync")
	fmt.Println("hello syncs")
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
			//SendRequest("localhost:10001", res)
			SendReply(chain.Client, res, port)
		} else {
			fmt.Println("inside sync" + chain.Next)
			SendRequest(chain.Next, res)
		}
	}
}

func startUDPService(port int, b *bank.Bank, chain *structs.Chain) {
	localAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}
	conn, err := net.ListenUDP("udp", &localAddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		buf := make([]byte, MAXLINE)
		n, _, err := conn.ReadFromUDP(buf)

		if err != nil {
			fmt.Printf("Error while reading UDP: %s\n", err)
		}

		rqst := &structs.Request{}
		json.Unmarshal(buf[:n], &rqst)

		reply := &structs.Request{}
		switch rqst.Transaction {
		case "getbalance":
			reply = b.GetBalance(rqst)
		case "withdraw":
			reply = b.Withdraw(rqst)
			fmt.Println("inside withdraw" + chain.Next)
		case "deposit":
			reply = b.Deposit(rqst)
			fmt.Println("inside deposit" + chain.Next)
		case "transfer":
			//TODO: phase 4
		}

		utils.Logoutput(chain.Server, reply.Requestid, reply.Outcome, reply.Balance, reply.Transaction)
		if chain.Istail {
			SendReply(chain.Client, reply, port)
		} else {
			SendRequest(chain.Next, reply)
		}

	}
}

func main() {
	b := bank.Initbank("wellsfargo", "wells")
	port, _ := strconv.Atoi(os.Args[1])
	utils.SetConfigFile("config.json")
	series, _ := strconv.Atoi(utils.Getconfig("chian1series"))
	lenservers, _ := strconv.Atoi(utils.Getconfig("chainlength"))
	curseries := int(port / 1000)
	series = series + (curseries - series)
	chain := structs.Makechain(series, port, lenservers)

	connMaster := connectToMaster(port)
	go sendOnlineMsg(connMaster)
	defer connMaster.Close()

	go startUDPService(port, b, chain)

	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		synchandler(w, r, b, chain, port)
	})
	err := http.ListenAndServe(chain.Server, nil)
	if err != nil {
		log.Fatal(err)
	}

}
