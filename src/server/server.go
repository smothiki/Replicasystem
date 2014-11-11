package main

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	//"io"
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

func connectToMaster() *net.UDPConn {
	ip, port := structs.GetIPAndPort(chain.Server)
	masterAddr := utils.Getconfig("master")
	localAddr := net.UDPAddr{
		Port: port + 100,
		IP:   net.ParseIP(ip),
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
	logMsg("SENT", request.String("HISTORY"))
	sent.PushBack(*request)
	fmt.Println("pushed", chain.Server)
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
	_, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR while sending ack", err)
	}
	client.Do(req)
	fmt.Println("SENT ack: " + ack.ReqKey)
	logMsg("SENT", "ack: "+ack.ReqKey)
}

/* SendReply sends reply to client */
func SendReply(request *structs.Request) {
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
	logMsg("SENT", request.String("REPLY"))
}

func synchandler(w http.ResponseWriter, r *http.Request, b *bank.Bank, port int) {
	fmt.Fprint(w, "Hello, sync")
	//fmt.Println("hello syncs")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		//fmt.Println(res)
		logMsg("RECV", res.String("HISTORY"))
		b.Set(res)
		utils.LogServer(chain.Server, res.Requestid, res.Account, res.Outcome, res.Transaction, res.Balance)
		sleepTime := rand.Intn(1500)
		fmt.Println("sleep for", sleepTime, "ms")
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		if chain.Istail {
			//time.Sleep(6000 * time.Millisecond)
			//SendRequest("localhost:10001", res)
			SendReply(res)
			ack := structs.Ack{
				ReqKey: res.MakeKey(),
			}
			SendAck(&ack)
		} else {
			SendRequest(res)
		}
	}
}

func alterChainHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprintf(w, "changed")
	body, _ := ioutil.ReadAll(r.Body)
	newChain := &structs.Chain{}
	json.Unmarshal(body, &newChain)
	logMsg("RECV", newChain.String())
	if chain.Prev != newChain.Prev && !newChain.Ishead {
		// if current server has new predecessor, send it the
		// last record in sent, and wait for sent records
		// after that entry
		sendLastSentToPrev(newChain.Prev, b)
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
		//old tail
		sendBankToTail(b, newChain)
		sendSentToTail(newChain)
	}
	chain = *newChain
}

func sendBankToTail(b *bank.Bank, newChain *structs.Chain) {
	//send basic bank info
	client := &http.Client{}
	newBank := bank.Initbank(b.Bankname, b.Bankid)
	msgBank, _ := json.Marshal(newBank)
	req, _ := http.NewRequest("POST", "http://"+newChain.Next+"/extend/bank", bytes.NewBuffer(msgBank))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	_, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("SENT Bank info: " + string(msgBank))
	logMsg("SENT", "Bank info: "+string(msgBank))

	//send accounts info
	acMap := b.Accounts()
	for _, pAcc := range *acMap {
		acc := *pAcc
		msg, _ := json.Marshal(acc)
		client := &http.Client{}
		req, _ := http.NewRequest("POST", "http://"+newChain.Next+"/extend/accounts", bytes.NewBuffer(msg))
		req.Header = http.Header{
			"accept": {"application/json"},
		}
		_, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("SENT a/c info: " + string(msg))
		logMsg("SENT", "a/c info: "+string(msg))
	}

	transMap := *b.TransMap()
	for _, trans := range transMap {
		msg, _ := json.Marshal(trans)
		client := &http.Client{}
		req, _ := http.NewRequest("POST", "http://"+newChain.Next+"/extend/transactions", bytes.NewBuffer(msg))
		req.Header = http.Header{
			"accept": {"application/json"},
		}
		_, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("SENT transaction: " + string(msg))
		logMsg("SENT", "transactions: "+string(msg))
	}
}

func sendSentToTail(newChain *structs.Chain) {
	lenSent := sent.Len()
	if lenSent > 0 {
		sentList := make([]structs.Request, lenSent)
		for e := sent.Front(); e != nil; e = e.Next() {
			r := e.Value.(structs.Request)
			sentList = append(sentList, r)
		}
		msg, _ := json.Marshal(sentList)
		client := &http.Client{}
		req, _ := http.NewRequest("POST", "http://"+newChain.Next+"/extend/sent", bytes.NewBuffer(msg))
		req.Header = http.Header{
			"accept": {"application/json"},
		}
		_, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("sent to tail:", sprtReqSlice(&sentList))
		logMsg("SENT", "sentlist: "+sprtReqSlice(&sentList))
	}
}

func extendBankHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprintf(w, "copied")
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &b)
	fmt.Println("bank info copied")
	logMsg("RECV", string(body))
	fmt.Println(b)
}

func extendAccountsHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprintf(w, "copied")
	body, _ := ioutil.ReadAll(r.Body)
	acc := bank.Account{}
	json.Unmarshal(body, &acc)
	b.AddAccount(acc.Accountid, acc.Balance)
	logMsg("RECV", string(body))
}

func extendTransactionsHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprintf(w, "copied")
	body, _ := ioutil.ReadAll(r.Body)
	var trans bank.Transaction
	json.Unmarshal(body, &trans)
	b.T.RecordTransaction(&trans)
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
	fmt.Println("RECV ack: " + ack.ReqKey)
	logMsg("RECV", "ack: "+ack.ReqKey)

	sleepTime := rand.Intn(1500)
	fmt.Println("sleep for", sleepTime, "ms")
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	SendAck(ack)
}

func requestSentHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	lastReq := &structs.Request{}
	json.Unmarshal(body, &lastReq)
	logMsg("RECV", lastReq.String("HISTORY"))
	fmt.Println("RECV", lastReq.String("HISTORY"))
	//l := list.New()
	var sendList []structs.Request

	if lastReq != nil {
		key := lastReq.MakeKey()
		bToAdd := false
		for e := sent.Front(); e != nil; e = e.Next() {
			req := e.Value.(structs.Request)
			if bToAdd {
				//l.PushBack(req)
				sendList = append(sendList, req)
			}
			if req.MakeKey() == key {
				bToAdd = true
			}
		}
	}

	enc := json.NewEncoder(w)
	enc.Encode(sendList)
	logMsg("SENT", "SentList: "+sprtReqSlice(&sendList))
	fmt.Println("SENT", "Sentlist: "+sprtReqSlice(&sendList))
	//sendSentsToNext(&sendList)
}

func sendLastSentToPrev(destServer string, b *bank.Bank) {
	e := sent.Back()
	r := e.Value.(structs.Request)
	msg, err := json.Marshal(r)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://"+destServer+"/requestSent", bytes.NewBuffer(msg))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR", err)
	}
	logMsg("SENT", r.String("HISTORY"))
	fmt.Println("SENT", r.String("HISTORY"))

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var sentList []structs.Request
	json.Unmarshal(body, &sentList)
	logMsg("RECV", "SentList: "+sprtReqSlice(&sentList))
	fmt.Println("RECV Sentlist:", sprtReqSlice(&sentList))

	for _, req := range sentList {
		b.Set(&req)
		utils.LogServer(chain.Server, req.Requestid, req.Account, req.Outcome, req.Transaction, req.Balance)
		//isleepTime := rand.Intn(1500)
		//fmt.Println("sleep for", sleepTime, "ms")
		//time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		if chain.Istail {
			//time.Sleep(6000 * time.Millisecond)
			//SendRequest("localhost:10001", res)
			SendReply(&req)
			ack := structs.Ack{
				ReqKey: req.MakeKey(),
			}
			SendAck(&ack)
		} else {
			SendRequest(&req)
		}
	}
}

/*func sendSentsToNext(lst *[]structs.Request) {
	msg, err := json.Marshal(lst)
	fmt.Println("msg", string(msg))
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://"+chain.Next+"/sent", bytes.NewBuffer(msg))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	_, err = client.Do(req)
	if err != nil {
		log.Println("ERROR while sending Sents to next server", chain.Next)
	}

	logMsg("SENT", "Sents "+string(msg))
}
*/

func startUDPService(b *bank.Bank) {
	ip, port := structs.GetIPAndPort(chain.Server)
	localAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
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
		logMsg("RECV", rqst.String("REQUEST"))

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
		//fmt.Println("dd", reply)
		utils.LogServer(chain.Server, reply.Requestid, reply.Account, reply.Outcome, reply.Transaction, reply.Balance)
		if chain.Istail {
			SendReply(reply)
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

	connMaster := connectToMaster()
	go sendOnlineMsg(connMaster)
	defer connMaster.Close()

	go startUDPService(b)

	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		synchandler(w, r, b, port)
	})
	http.HandleFunc("/alterChain", func(w http.ResponseWriter, r *http.Request) {
		alterChainHandler(w, r, b)
	})

	http.HandleFunc("/ack", func(w http.ResponseWriter, r *http.Request) {
		ackHandler(w, r)
	})
	http.HandleFunc("/extendChain", func(w http.ResponseWriter, r *http.Request) {
		extendChainHandler(w, r, b)
	})

	http.HandleFunc("/extend/bank", func(w http.ResponseWriter, r *http.Request) {
		extendBankHandler(w, r, b)
	})
	http.HandleFunc("/extend/accounts", func(w http.ResponseWriter, r *http.Request) {
		extendAccountsHandler(w, r, b)
	})

	http.HandleFunc("/extend/transactions", func(w http.ResponseWriter, r *http.Request) {
		extendTransactionsHandler(w, r, b)
	})
	http.HandleFunc("/requestSent", requestSentHandler)
	err := http.ListenAndServe(chain.Server, nil)
	if err != nil {
		log.Fatal(err)
	}

}

func sprtReqSlice(rs *[]structs.Request) string {
	r := "["
	l := len(*rs)
	for idx, req := range *rs {
		r += req.String("HISTORY")
		if idx < l-1 {
			r += ","
		}
	}
	r += "]"
	return r
}
