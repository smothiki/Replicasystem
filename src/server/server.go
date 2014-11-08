package main

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
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

var sent list.List
var chain structs.Chain
var recvNum, sendNum int

func logMsg(msgType, msg string) {
	if msgType == "SENT" {
		utils.LogSMsg(chain.Server, msgType, sendNum, msg)
		sendNum++
	} else if msgType == "RECV" {
		utils.LogSMsg(chain.Server, msgType, recvNum, msg)
		recvNum++
	} else {
		log.Println("LOG ERROR: UNKOWN MSG TYPE")
	}
}

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
		log.Println("ERROR while connecting to master.", err)
	}
	utils.LogSEvent(chain.Server, "Connected to master")
	return conn
}

func sendOnlineMsg(conn *net.UDPConn) {
	for {
		msg, _ := json.Marshal(1)
		_, err := conn.Write(msg)
		logMsg("SENT", "ONLINE")
		if err != nil {
			log.Println("ERROR while sending online msg to master.", err)
		}
		time.Sleep(SENT_HEALTH_INTERVAL * time.Millisecond)
	}
}

/* SendRequest send request to successor */
func SendRequest(request *structs.Request) {
	res1B, err := json.Marshal(request)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://"+chain.Next+"/sync", bytes.NewBuffer(res1B))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	_, err = client.Do(req)
	if err != nil {
		log.Println("Error while sending request.", err)
	}
	logMsg("SENT", request.String())
	sent.PushBack(request)
}

func SendAck(ack *structs.Ack) {
	if chain.Ishead {
		return
	}
	msg, _ := json.Marshal(ack)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://"+chain.Prev+"/ack", bytes.NewBuffer(msg))
	req.Close = true
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	/*_, err = client.Do(req)
	if err != nil {
		fmt.Println("ERROR while sending ack", err)
	}*/
	client.Do(req)
	//logMsg("SENT", ack)
}

/* SendReply sends reply to client */
func SendReply(request *structs.Request, port int) {
	res1B, err := json.Marshal(request)
	/*localAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}*/
	log.Println("cient addr", request.Client)
	conn, err := net.DialUDP("udp", nil, &request.Client)

	if err != nil {
		log.Println("ERROR while connecting to client", err)
	}

	defer conn.Close()

	_, err = conn.Write(res1B)
	if err != nil {
		log.Println("ERROR while sending reply to client", err)
	}
	logMsg("SENT", request.String())
}

func synchandler(w http.ResponseWriter, r *http.Request, b *bank.Bank, port int) {
	fmt.Fprint(w, "Hello, sync")
	//fmt.Println("hello syncs")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		//fmt.Println(res)
		logMsg("RECV", res.String())
		b.Set(res)
		utils.LogServer(chain.Server, res.Requestid, res.Account, res.Outcome, res.Transaction, res.Balance)
		sleepTime := rand.Intn(1500)
		fmt.Println("sleep for", sleepTime, "ms")
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		if chain.Istail {
			//time.Sleep(6000 * time.Millisecond)
			//SendRequest("localhost:10001", res)
			SendReply(res, port)
			ack := structs.Ack{
				ReqKey: res.MakeKey(),
			}
			SendAck(&ack)
		} else {
			SendRequest(res)
		}
	}
}

func alterChainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "changed")
	body, _ := ioutil.ReadAll(r.Body)
	newChain := &structs.Chain{}
	json.Unmarshal(body, &newChain)
	logMsg("RECV", newChain.String())
	if chain.Prev != newChain.Prev && !newChain.Ishead {
		// if current server has new predecessor, send it the
		// last record in sent, and wait for sent records
		// after that entry
		sendLastSentToPrev(newChain.Prev)
	}
	//if chain.Next != newChain.Next && !newChain.Istail {}
	chain = *newChain
}

func extendChainHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprintf(w, "extended")
	body, _ := ioutil.ReadAll(r.Body)
	newChain := &structs.Chain{}
	json.Unmarshal(body, &newChain)
	logMsg("RECV", newChain.String())
	if chain.Istail && !newChain.Istail {
		msg, _ := json.Marshal(b)
		client := &http.Client{}
		req, _ := http.NewRequest("POST", "http://"+newChain.Next+"/copyBank", bytes.NewBuffer(msg))
		req.Header = http.Header{
			"accept": {"application/json"},
		}
		_, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		logMsg("SENT", string(msg))
	}
	chain = *newChain
}

func copyBankHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprintf(w, "copied")
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &b)
	log.Println("bank info copied")
	logMsg("RECV", string(body))
}

func ackHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "r")
	body, _ := ioutil.ReadAll(r.Body)
	ack := &structs.Ack{}
	json.Unmarshal(body, &ack)
	for e := sent.Front(); e != nil; e = e.Next() {
		req := e.Value.(structs.Request)
		if (req).MakeKey() == ack.ReqKey {
			sent.Remove(e)
			break
		}
	}
	SendAck(ack)
}

func requestSentHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	elem := &list.Element{}
	json.Unmarshal(body, &elem)
	lastRec := elem.Value.(structs.Request)
	logMsg("RECV", lastRec.String())
	l := list.New()

	if elem != nil {
		key := lastRec.MakeKey()
		bToAdd := false
		for e := sent.Front(); e != nil; e = e.Next() {
			if bToAdd {
				l.PushBack(e)
			}
			req := e.Value.(structs.Request)
			if req.MakeKey() == key {
				bToAdd = true
			}
		}
	}
	sendSentsToNext(l)
}

func sendLastSentToPrev(destServer string) {
	e := sent.Back()
	msg, err := json.Marshal(e)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://"+destServer+"/requestSent", bytes.NewBuffer(msg))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	_, err = client.Do(req)
	if err != nil {
		log.Println("ERROR", err)
	}
	logMsg("SENT", e.Value.(string))
}

func sendSentsToNext(lst *list.List) {
	for e := lst.Front(); e != nil; e = e.Next() {
		req := e.Value.(structs.Request)
		SendRequest(&req)
		logMsg("SENT", req.String())
	}
}

func startUDPService(port int, b *bank.Bank) {
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
		logMsg("RECV", rqst.String())

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

		reply.Client = rqst.Client
		reply.Time = rqst.Time
		fmt.Println("dd", reply)
		utils.LogServer(chain.Server, reply.Requestid, reply.Account, reply.Outcome, reply.Transaction, reply.Balance)
		if chain.Istail {
			SendReply(reply, port)
		} else {
			SendRequest(reply)
		}

	}
}

func die() {
	utils.LogSEvent(chain.Server, "Server died!")
	os.Exit(0)
}

func main() {
	recvNum = 0
	sendNum = 0
	b := bank.Initbank("wellsfargo", "wells")
	port, _ := strconv.Atoi(os.Args[1])
	utils.SetConfigFile(os.Args[2])
	series, _ := strconv.Atoi(utils.Getconfig("chian1series"))
	lenservers, _ := strconv.Atoi(utils.Getconfig("chainlength"))
	curseries := int(port / 1000)
	series = series + (curseries - series)
	chain = *structs.Makechain(series, port, lenservers)
	lifetime := utils.GetLifeTime(port%1000 - 1)
	startDelay := utils.GetStartDelay(port%1000 - 1)
	if startDelay != 0 {
		time.Sleep(time.Duration(startDelay*1000) * time.Millisecond)
	}
	utils.LogSEvent(chain.Server, "Server started! "+chain.String())

	if lifetime != 0 {
		utils.SetTimer(lifetime, die)
	}

	connMaster := connectToMaster(port)
	go sendOnlineMsg(connMaster)
	defer connMaster.Close()

	go startUDPService(port, b)

	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		synchandler(w, r, b, port)
	})
	http.HandleFunc("/alterChain", func(w http.ResponseWriter, r *http.Request) {
		alterChainHandler(w, r)
	})

	http.HandleFunc("/ack", func(w http.ResponseWriter, r *http.Request) {
		ackHandler(w, r)
	})
	http.HandleFunc("/extendChain", func(w http.ResponseWriter, r *http.Request) {
		extendChainHandler(w, r, b)
	})

	http.HandleFunc("/copyBank", func(w http.ResponseWriter, r *http.Request) {
		copyBankHandler(w, r, b)
	})
	http.HandleFunc("/requestSent", requestSentHandler)
	err := http.ListenAndServe(chain.Server, nil)
	if err != nil {
		log.Fatal(err)
	}

}
