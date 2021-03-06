package main

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"strconv"
	//"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"

	"time"

	bank "github.com/replicasystem/src/commons/bank"
	"github.com/replicasystem/src/commons/structs"
	"github.com/replicasystem/src/commons/utils"
)

const MAXLINE = 1024                 // max length of char buffer
var sendOnlineCycle time.Duration    // frequency of sending online msg to master
var ackProcMaxTime int               // time of (simulated) ack processing
var rqstProcMaxTime int              // time of (simulated) request processing
var extendSendInterval time.Duration // interval of sending histories to new tail during extension
var checkOnlineCycle int             //frequency of master checking online msg

var sent list.List                // 'Sent'
var chain structs.Chain           // info of current server
var recvNum, sendNum, lossNum int // msg counters for logging
var lossProb float32              //probability of message loss

//logMsg logs sent/received message msg into log file,
//msgType is "SENT" or "RECV", counterServer is receiving / sender
func logMsg(msgType, msg, counterServer string) {
	if msgType == "SENT" {
		msg += " (to " + counterServer + ")"
		utils.LogSMsg(chain.Server, msgType, sendNum, msg)
		sendNum++
	} else if msgType == "RECV" {
		msg += " (from " + counterServer + ")"
		utils.LogSMsg(chain.Server, msgType, recvNum, msg)
		recvNum++
	} else if msgType == "LOSS" {
		utils.LogSMsg(chain.Server, msgType, lossNum, msg)
		lossNum++
	} else {
		log.Println("LOG ERROR: UNKOWN MSG TYPE")
	}
}

//connectToMaster connects UDP socket to master server,
//returning socket descriptor
func connectToMaster() *net.UDPConn {
	ip, port := utils.GetIPAndPort(chain.Server)
	masterAddr := utils.Getconfig("master")
	localAddr := net.UDPAddr{
		Port: port + 100,
		IP:   net.ParseIP(ip),
	}
	destIP, destPort := utils.GetIPAndPort(masterAddr)
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

//sendOnlingMsg sends online messge (for health check) to master
//server via conn returned by connectToMaster
func sendOnlineMsg(conn *net.UDPConn) {
	for {
		msg, _ := json.Marshal(1)
		_, err := conn.Write(msg)
		if chain.Available {
			logMsg("SENT", "ONLINE", "master")
		}
		if err != nil {
			log.Println("ERROR while sending online msg to master.", err)
		}
		time.Sleep(sendOnlineCycle * time.Millisecond)
	}
}

//SendRequest send request to successor
func SendRequest(request *structs.Request, dest string) {
	randomSleep(rqstProcMaxTime, "before sending request")
	res1B, err := json.Marshal(request)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://"+dest+"/sync", bytes.NewBuffer(res1B))
	req.Header = http.Header{
		"accept": {"application/json"},
	}

	var format string
	if request.Transaction == "transfer" {
		format = "TRANS_HIST"
	} else {
		format = "HISTORY"
	}
	logMsg("SENT", request.String(format), dest)
	sent.PushBack(*request)
	utils.LogSEvent(chain.Server, "Added "+request.MakeKey()+" into 'Sent'")

	_, err = client.Do(req)
	if err != nil {
		log.Println("Error while sending request.", err)
	}
}

//SendAck sends acknowledgement ack to predecessor
func SendAck(ack *structs.Ack) {
	if chain.Ishead || chain.Prev == "" {
		return
	}
	randomSleep(ackProcMaxTime, "before sending ack")
	msg, _ := json.Marshal(ack)
	client := &http.Client{}
	if chain.Ishead || chain.Prev == "" {
		return
	}
	req, _ := http.NewRequest("POST", "http://"+chain.Prev+"/ack", bytes.NewBuffer(msg))
	req.Close = true
	req.Header = http.Header{
		"accept": {"application/json"},
	}

	fmt.Println("SENT ack: " + ack.ReqKey)
	logMsg("SENT", "ack: "+ack.ReqKey, chain.Prev)

	_, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR while sending ack", err)
	}
}

//SendReply sends reply (request) to request.Receiver
func SendReply(request *structs.Request) {
	if rand.Float32() < lossProb {
		logMsg("LOSS", request.String("REPLY"), request.Receiver.String())
		return
	}
	randomSleep(rqstProcMaxTime, "before sending request")
	res1B, err := json.Marshal(request)
	/*localAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}*/
	log.Println("cient addr", request.Receiver)
	conn, err := net.DialUDP("udp", nil, &request.Receiver)

	if err != nil {
		log.Println("ERROR while connecting to client", err)
	}

	defer conn.Close()

	_, err = conn.Write(res1B)
	if err != nil {
		log.Println("ERROR while sending reply to client", err)
	}
	logMsg("SENT", request.String("REPLY"), request.Receiver.String())
}

//sync handler processes request sent by predecessor and sends it
//to successor server
func synchandler(w http.ResponseWriter, r *http.Request, b *bank.Bank, port int) {
	fmt.Fprint(w, "Hello, sync")
	if r.Method == "POST" {
		body, _ := ioutil.ReadAll(r.Body)
		res := &structs.Request{}
		json.Unmarshal(body, &res)
		var format string
		if res.Transaction == "transfer" {
			format = "TRANS_HIST"
		} else {
			format = "HISTORY"
		}
		logMsg("RECV", res.String(format), chain.Prev)
		b.Set(res)
		utils.LogEventData(chain.Server, "server", "PROC", res.String("REPLY"))
		if chain.Istail {
			if res.Transaction == "transfer" {
				ip, port := utils.GetIPAndPort(chain.Server)
				res.Sender = net.UDPAddr{
					Port: port,
					IP:   net.ParseIP(ip),
				}
				if res.DestBank == b.Bankid {
					SendReply(res)
				} else {
					if res.Outcome == "processed" {
						dest := queryDestBankHead(res.DestBank)
						sendTransferToDest(res, dest)
						if chain.FailOnSendTrans {
							utils.LogSEvent(chain.Server, "Server failed on sending request to dest bank")
							os.Exit(0)
						}
						if isFailOnRecvTrans(dest) {
							var newdest string
							time.Sleep(time.Duration(checkOnlineCycle+1) * time.Second)
							newdest = queryDestBankHead(res.DestBank)
							utils.LogSEvent(chain.Server, "dest bank head failed, retransmitting transfer request...")
							sendTransferToDest(res, newdest)
						}
					} else {
						res.Receiver = res.Client
						SendReply(res)
						ack := structs.Ack{
							ReqKey: res.MakeKey(),
						}
						SendAck(&ack)
					}
					return
				}
			} else {
				SendReply(res)
			}

			ack := structs.Ack{
				ReqKey: res.MakeKey(),
			}
			SendAck(&ack)
		} else {
			SendRequest(res, chain.Next)
		}
	}
}

//alterChainHandler handles http request "alterChain", which
//indicates predecessor or successor has changed
func alterChainHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprintf(w, "changed")
	body, _ := ioutil.ReadAll(r.Body)
	newChain := &structs.Chain{}
	json.Unmarshal(body, &newChain)
	logMsg("RECV", newChain.String(), "master")
	hasNewPrev := chain.Prev != newChain.Prev && !newChain.Ishead
	isNewTail := !chain.Istail && newChain.Istail
	chain.SetChain(newChain)
	if hasNewPrev {
		// if current server has new predecessor, send it the
		// last record in sent, and wait for sent records
		// after that entry
		sendLastSentToPrev(newChain.Prev, b)
	}
	//if chain.Next != newChain.Next && !newChain.Istail {}
	if isNewTail {
		time.Sleep(2 * time.Second)
		e := sent.Front()
		for e != nil {
			r := e.Value.(structs.Request)
			if r.Transaction == "transfer" && b.Bankid != r.DestBank {
				utils.LogSEvent(chain.Server, "Sending entry in 'Sent' to dest bank")
				dest := queryDestBankHead(r.DestBank)
				ip, port := utils.GetIPAndPort(chain.Server)
				r.Sender = net.UDPAddr{
					Port: port,
					IP:   net.ParseIP(ip),
				}
				sendTransferToDest(&r, dest)
				e = e.Next()
			} else {
				utils.LogSEvent(chain.Server, "Removing entry in 'Sent' and send ack")
				ack := structs.Ack{
					ReqKey: r.MakeKey(),
				}
				SendAck(&ack)
				sent.Remove(e)
				e = sent.Front()
			}
		}
	}
}

//extendChainHandler handles http request "extendChain", which
//indicates new tail being added to chain. Both old and new tail
//show deal with extendChain request
func extendChainHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	body, _ := ioutil.ReadAll(r.Body)
	newChain := &structs.Chain{}
	json.Unmarshal(body, &newChain)
	logMsg("RECV", newChain.String(), "master")
	if chain.Istail && !newChain.Istail {
		//old tail
		sendBankToTail(b, newChain)
		sendSentToTail(newChain)
		fmt.Fprintf(w, "extended")
	} else {
		//new tail
		if chain.FailOnExtension {
			utils.LogSEvent(chain.Server, "Failed during extension")
			chain.Available = false
			fmt.Fprintf(w, "failed")
		} else {
			fmt.Fprintf(w, "extended")
		}
	}
	chain.SetChain(newChain)
}

//sendBankToTail sends bank information in old tail to new tail
func sendBankToTail(b *bank.Bank, newChain *structs.Chain) {
	//send basic bank info
	client := &http.Client{}
	newBank := bank.Initbank(b.Bankname, b.Bankid)
	msgBank, _ := json.Marshal(newBank)
	req, _ := http.NewRequest("POST", "http://"+newChain.Next+"/extend/bank", bytes.NewBuffer(msgBank))
	req.Header = http.Header{
		"accept": {"application/json"},
	}

	fmt.Println("SENT Bank info: " + string(msgBank))
	logMsg("SENT", "Bank info: "+string(msgBank), newChain.Next)

	_, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	time.Sleep(extendSendInterval * time.Millisecond)

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

		fmt.Println("SENT a/c info: " + string(msg))
		logMsg("SENT", "a/c info: "+string(msg), newChain.Next)

		_, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(extendSendInterval * time.Millisecond)
	}

	//send transactions
	transMap := *b.TransMap()
	for _, trans := range transMap {
		msg, _ := json.Marshal(trans)
		client := &http.Client{}
		req, _ := http.NewRequest("POST", "http://"+newChain.Next+"/extend/transactions", bytes.NewBuffer(msg))
		req.Header = http.Header{
			"accept": {"application/json"},
		}

		fmt.Println("SENT transaction: " + string(msg))
		logMsg("SENT", "transaction: "+string(msg), newChain.Next)

		_, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(extendSendInterval * time.Millisecond)
	}
}

//sendSentToTail sends 'Sent' in old tail to new tail
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

		fmt.Println("sent to tail:", sprtReqSlice(&sentList))
		logMsg("SENT", "'Sent': "+sprtReqSlice(&sentList), newChain.Next)

		_, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
	} else {
		utils.LogSEvent(chain.Server, "'Sent' is empty, nothing to send to new tail")
	}
}

//extendBankHandler handles http request "extend/bank", new tail
//calls this function to receive basic bank information sent by
//old tail
func extendBankHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprintf(w, "copied")
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &b)
	fmt.Println("bank info copied")
	logMsg("RECV", string(body), chain.Prev)
	fmt.Println(b)
}

//extendAccountsHandler handles http request "extend/accounts", new
//tail calls this function to receive accounts information sent by
//old tail
func extendAccountsHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprintf(w, "copied")
	body, _ := ioutil.ReadAll(r.Body)
	acc := bank.Account{}
	json.Unmarshal(body, &acc)
	b.AddAccount(acc.Accountid, acc.Balance)
	logMsg("RECV", string(body), chain.Prev)
}

//extendTransactionsHandler handles http request "extend/transactions",
//new tail calls this function to receive transaction history sent by
//old tail
func extendTransactionsHandler(w http.ResponseWriter, r *http.Request, b *bank.Bank) {
	fmt.Fprintf(w, "copied")
	body, _ := ioutil.ReadAll(r.Body)
	var trans bank.Transaction
	json.Unmarshal(body, &trans)
	b.T.RecordTransaction(&trans)
	logMsg("RECV", string(body), chain.Prev)
}

//ackHandler handles http request "ack" sent by its successor,
//removing corresponding request from 'Sent' and sending ack
//to predecessor
func ackHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "r")
	body, _ := ioutil.ReadAll(r.Body)
	ack := &structs.Ack{}
	json.Unmarshal(body, &ack)
	for e := sent.Front(); e != nil; e = e.Next() {
		req := e.Value.(structs.Request)
		if req.MakeKey() == ack.ReqKey {
			utils.LogSEvent(chain.Server, "Removed "+req.MakeKey()+" from 'Sent'")
			sent.Remove(e)
			break
		}
	}
	logMsg("RECV", "ack: "+ack.ReqKey, chain.Next)
	SendAck(ack)
}

//requestSentHandler handles http request "requestSent" sent by new
//successor. When successor fails, new successor sends its last entry
//in 'Sent'. On receiving it, it sends back those records in its
//'Sent' but not in new successor's 'Sent'.
func requestSentHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	lastReq := &structs.Request{}
	json.Unmarshal(body, &lastReq)
	logMsg("RECV", "Last entry in 'Sent': "+lastReq.String("HISTORY"), "new successor")
	fmt.Println("RECV", "Last entry in 'Sent': "+lastReq.String("HISTORY"))
	var sendList []structs.Request

	if lastReq != nil {
		key := lastReq.MakeKey()
		bToAdd := false
		if lastReq.Requestid == "" {
			bToAdd = true
		}
		for e := sent.Front(); e != nil; e = e.Next() {
			req := e.Value.(structs.Request)
			if bToAdd {
				sendList = append(sendList, req)
			} else if req.MakeKey() == key {
				bToAdd = true
			}
		}
	}

	if chain.FailOnReqSent {
		utils.LogSEvent(chain.Server, "Failed after receiving last entry in 'Sent'")
		os.Exit(0)
	}

	enc := json.NewEncoder(w)
	enc.Encode(sendList)
	logMsg("SENT", "'Sent': "+sprtReqSlice(&sendList), "new sucessor")
	fmt.Println("SENT", "'Sent': "+sprtReqSlice(&sendList))
}

//sendLastSentToPrev sends last entry of 'Sent' to its new predecessor
//destServer and waits for new 'Sent' records and processes them
func sendLastSentToPrev(destServer string, b *bank.Bank) {
	r := structs.Request{}
	e := sent.Back()
	if e != nil {
		r = e.Value.(structs.Request)
	}
	msg, err := json.Marshal(r)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://"+destServer+"/requestSent", bytes.NewBuffer(msg))
	req.Header = http.Header{
		"accept": {"application/json"},
	}

	logMsg("SENT", "Last entry in 'Sent': "+r.String("HISTORY"), destServer)
	fmt.Println("SENT", "Last entry in 'Sent': "+r.String("HISTORY"))

	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR", err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var sentList []structs.Request
	json.Unmarshal(body, &sentList)
	logMsg("RECV", "'Sent': "+sprtReqSlice(&sentList), destServer)
	fmt.Println("RECV 'Sent': ", sprtReqSlice(&sentList))
	if chain.FailOnRecvSent {
		utils.LogSEvent(chain.Server, "Failed on receiving 'Sent' from predecessor")
		os.Exit(0)
	}

	for _, req := range sentList {
		b.Set(&req)
		utils.LogEventData(chain.Server, "server", "PROC", req.String("REPLY"))
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
			SendRequest(&req, chain.Next)
		}
	}
}

//startUDPService listens UDP socket, through which clients send
//request, and process the incoming requests, and then sends
//requests either to clients or successor server
func startUDPService(b *bank.Bank) {
	ip, port := utils.GetIPAndPort(chain.Server)
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
		if rand.Float32() < lossProb {
			logMsg("LOSS", rqst.String("REQUEST"), rqst.Receiver.String())
			//TODO
			continue
		}
		if rqst.Transaction == "transfer" {
			var sender string
			if rqst.Sender.String() == "0.0.0.0:0" {
				sender = rqst.Receiver.String()
			} else {
				sender = rqst.Sender.String()
			}
			logMsg("RECV", rqst.String("TRANS_REQ"), sender)
		} else {
			logMsg("RECV", rqst.String("REQUEST"), rqst.Receiver.String())
		}

		reply := &structs.Request{}
		switch rqst.Transaction {
		case "getbalance":
			reply = b.GetBalance(rqst)
			reply.Receiver = rqst.Receiver
		case "withdraw":
			reply = b.Withdraw(rqst)
			reply.Receiver = rqst.Receiver
		case "deposit":
			reply = b.Deposit(rqst)
			reply.Receiver = rqst.Receiver
		case "transfer":
			//transfer operation will check the dst bank and perform the necessary action
			//if current bank is not the dst bank it will withdraw the amount else deposit the amount
			if rqst.DestBank != b.Bankid &&
				rqst.Receiver.String() != chain.Server {
				//initially Receiver is client
				//srcBank received request sent by client
				reply = b.Transfer(rqst)
				fmt.Println("reply result", reply)
				reply.Time = rqst.Time
				reply.Client = rqst.Client
				if reply.Outcome == "processed" {
					ip, port := utils.GetIPAndPort(chain.Server)
					rqst.Sender = net.UDPAddr{
						Port: port,
						IP:   net.ParseIP(ip),
					}
					utils.LogEventData(chain.Server, "server", "PROC", reply.String("REPLY"))
					if chain.Istail {
						if reply.Outcome == "processed" {
							dest := queryDestBankHead(rqst.DestBank)
							sendTransferToDest(rqst, dest)
							if isFailOnRecvTrans(dest) {
								var newdest string
								time.Sleep(time.Duration(checkOnlineCycle+1) * time.Second)
								newdest = queryDestBankHead(rqst.DestBank)
								utils.LogSEvent(chain.Server, "dest bank head failed, retransmitting transfer request...")
								sendTransferToDest(rqst, newdest)
							} else {
								reply.Receiver = reply.Client
								SendReply(reply)
							}
						}
					} else {
						SendRequest(reply, chain.Next)
					}
					continue
				}
			} else if rqst.DestBank == b.Bankid {
				//destBank received rqst sent by srcBank
				reply = b.Transfer(rqst)
				//make "client" src bank head server
				reply.Client = rqst.Client
				reply.Receiver = rqst.Sender
				ip, port := utils.GetIPAndPort(chain.Server)
				reply.Sender = net.UDPAddr{
					Port: port,
					IP:   net.ParseIP(ip),
				}
				if chain.FailOnRecvTrans {
					utils.LogSEvent(chain.Server, "Failed on receiving transfer request from source bank.")
					os.Exit(0)
				}
			} else if rqst.Receiver.String() == chain.Server {
				//srcBank received reply from destBank
				reply = b.Transfer(rqst)
				reply.Client = rqst.Client
				reply.Receiver = rqst.Client
				reply.Time = rqst.Time
				if chain.Istail && !chain.Ishead {
					ack := structs.Ack{
						ReqKey: reply.MakeKey(),
					}
					SendAck(&ack)
				}
			}
		}

		reply.Time = rqst.Time
		utils.LogEventData(chain.Server, "server", "PROC", reply.String("REPLY"))
		if chain.Istail {
			SendReply(reply)
		} else {
			SendRequest(reply, chain.Next)
		}
	}
}

//sendTransferToDest sends transfer request (by source tail) to dest (dest head)
func sendTransferToDest(request *structs.Request, dest string) {
	msg, _ := json.Marshal(request)
	ip, port := utils.GetIPAndPort(dest)
	remoteAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}
	ip, port = utils.GetIPAndPort(chain.Server)
	localAddr := net.UDPAddr{
		Port: port + 150,
		IP:   net.ParseIP(ip),
	}
	ip, port = utils.GetIPAndPort(chain.Server)
	conn, err := net.DialUDP("udp", &localAddr, &remoteAddr)
	if err != nil {
		log.Println("ERROR while connecting to transfer dest bank.")
	}
	defer conn.Close()

	_, err = conn.Write(msg)
	if err != nil {
		log.Println("Error while sending transfer request to dest bank")
	}
	logMsg("SENT", request.String("TRANS_REQ"), dest)
}

func isFailOnRecvTrans(addr string) bool {
	_, port := utils.GetIPAndPort(addr)
	return utils.GetFailOnRecvTrans(port%1000 - 1)
}

//queryDestBankHead queries from masterhead server address of
//the destination bank during transfer
func queryDestBankHead(destBank string) string {
	client := &http.Client{}
	master := utils.Getconfig("master")
	rqstHead := structs.DestHeadRqst{
		DestBank: destBank,
		Sender:   chain.Server,
	}
	msg, _ := json.Marshal(&rqstHead)
	req, _ := http.NewRequest("POST", "http://"+master+"/transfer/destHead", bytes.NewBuffer(msg))
	req.Header = http.Header{
		"accept": {"text/plain"},
	}
	req.Close = true

	fmt.Println("SENT Query dest bank head", destBank)
	logMsg("SENT", "Query Bank "+destBank+" head server", master)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR", err)
	}

	//defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	result := string(body)
	resp.Body.Close()
	fmt.Println("RECV dest head", result)
	logMsg("RECV", "Head server of bank "+destBank+": "+result, "master")
	return result
}

//die terminate current server to simulate server failure
func die() {
	utils.LogSEvent(chain.Server, "Server died!")
	os.Exit(0)
}

func main() {
	//read config
	recvNum = 0
	sendNum = 0
	lossNum = 0
	port, _ := strconv.Atoi(os.Args[1])
	utils.SetConfigFile(os.Args[2])
	series := utils.GetConfigInt("chain1series")
	lenservers := utils.GetConfigInt("chainlength")
	curseries := int(port / 1000)

	//bank name and bank ID will have th current chain series to identify unique bank

	b := bank.Initbank("wells", strconv.Itoa(curseries))
	series = series + (curseries - series)
	chain = *structs.Makechain(series, port, lenservers)
	chain.FailOnReqSent = utils.GetFailOnReqSent(port%1000 - 1)
	chain.FailOnRecvSent = utils.GetFailOnRecvSent(port%1000 - 1)
	chain.FailOnExtension = utils.GetFailOnExtension(port%1000 - 1)
	chain.FailOnRecvTrans = utils.GetFailOnRecvTrans(port%100 - 1)
	chain.FailOnSendTrans = utils.GetFailOnSendTrans(port%100 - 1)

	r, _ := strconv.ParseFloat(utils.Getconfig("msgLossProb"), 32)
	lossProb = float32(r)
	ackProcMaxTime = utils.GetConfigInt("ackProcMaxTime")
	rqstProcMaxTime = utils.GetConfigInt("rqstProcMaxTime")
	sendOnlineCycle = time.Duration(utils.GetConfigInt("sendOnlineCycle"))
	extendSendInterval = time.Duration(utils.GetConfigInt("extendSendInterval"))
	checkOnlineCycle = utils.GetConfigInt("checkOnlineCycle")

	lifetime := utils.GetLifeTime(port%1000 - 1)
	startDelay := utils.GetStartDelay(port%1000 - 1)
	if startDelay != 0 {
		time.Sleep(time.Duration(startDelay*1000) * time.Millisecond)
	}

	//service startup
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

//sprtReqSlice conversts slice of requests rs to string
func sprtReqSlice(rs *[]structs.Request) string {
	r := "["
	l := len(*rs)
	for idx, req := range *rs {
		r += "{" + req.String("HISTORY") + "}"
		if idx < l-1 {
			r += ", "
		}
	}
	r += "]"
	return r
}

//randomSleep sleeps for random duration, whose upperbound is
//upperTime seconds and prints out msg
func randomSleep(upperTime int, msg string) {
	sleepTime := rand.Intn(upperTime)
	fmt.Println("sleep for", sleepTime, "ms", msg)
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
}
