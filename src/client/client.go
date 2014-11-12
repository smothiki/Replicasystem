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

const MAXLINE = 1024 //max size of char buffer

var chain structs.Chain  //info of chain
var recvNum, sendNum int //msg counter for logging

//logMsg logs msg to log file, msgType can be "SENT"
//or "RECV", counterServer is corresponding receiver
//or sender
func logMsg(msgType, msg, counterServer string) {
	if msgType == "SENT" {
		msg += " (to " + counterServer + ")"
		utils.LogCMsg(chain.Server, msgType, sendNum, msg)
		sendNum++
	} else if msgType == "RECV" {
		msg += " (from " + counterServer + ")"
		utils.LogCMsg(chain.Server, msgType, recvNum, msg)
		recvNum++
	} else {
		log.Println("LOG ERROR: UNKOWN MSG TYPE")
	}
}

//SendRequest sends request (query/update) to server
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
	request.Time = fmt.Sprintf("%d", (time.Now().Unix()))
	res1B, err := json.Marshal(request)

	_, err = conn.Write(res1B)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("SENT", request.String("REQUEST"))
	logMsg("SENT", request.String("REQUEST"), server)
}

//createUDPSocket creates and listen UDP socket, through
//which servers send responses
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
	utils.LogCEvent(chain.Server, "UDP Socket connected!")
	return conn
}

//readResponses read response and returns *structs.Request
//from conn sent by servers
func readResponse(conn *net.UDPConn) *structs.Request {
	buf := make([]byte, MAXLINE)
	n, _, err := conn.ReadFromUDP(buf)

	if err != nil {
		fmt.Println("Error while reading UDP", err)
	}

	rqst := &structs.Request{}
	json.Unmarshal(buf[:n], &rqst)
	fmt.Println("RECV", rqst.String("REPLY"))
	logMsg("RECV", rqst.String("REPLY"), "SERVER")

	return rqst
}

//alterChainHandler handles change of chain by setting
//new header or tail
func alterChainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "changed")
	body, _ := ioutil.ReadAll(r.Body)
	newHeadTail := &structs.ClientNotify{}
	json.Unmarshal(body, &newHeadTail)
	if newHeadTail.Head != "" {
		chain.Head = newHeadTail.Head
		fmt.Println("newHead", chain.Head)
		logMsg("RECV", "New head is "+chain.Head, r.RemoteAddr)
	} else if newHeadTail.Tail != "" {
		chain.Tail = newHeadTail.Tail
		fmt.Println("newTail", chain.Tail)
		logMsg("RECV", "New tail is "+chain.Tail, r.RemoteAddr)
	}
}

//simulate simulates the client and sends request to servers
func simulate(conn *net.UDPConn, port, clientIdx int, waitTime time.Duration) {
	//reqGenMethod := utils.GetTestCaseGenMethod(clientIdx)
	reqFile := utils.GetTestRequestFile(clientIdx)
	/*var listreqs *[]structs.Request
	if reqGenMethod == "predefined" {
		listreqs = structs.GetTestreqs(reqFile)
	} else {
		listreqs = structs.GetrequestList(3, "getbalance", reqFile)
	}*/
	listreqs := structs.GenRequestList(reqFile)
	var dest string
	//xxx := 0
	for _, request := range *listreqs {
		if request.Transaction == "getbalance" {
			dest = chain.Tail
		} else {
			dest = chain.Head
		}

		err := utils.Timeout("Request", waitTime*time.Millisecond,
			func() {
				//fmt.Println("ENTER LOOP", xxx)
				SendRequest(dest, &request, port)
				//fmt.Println("MID   LOOP", xxx)
				fmt.Println("result", *readResponse(conn))
				//fmt.Println("LEAVE LOOP", xxx)
				fmt.Println()
				return
			} /*, xxx*/)
		if err != nil {
			//fmt.Println(xxx, err)
		}
		//xxx++
	}
}

func main() {
	//read configuration
	port, _ := strconv.Atoi(os.Args[1])
	utils.SetConfigFile(os.Args[2])
	series := utils.GetConfigInt("chain1series")
	lenservers := utils.GetConfigInt("chainlength")
	curseries := int(port / 1000)
	series = series + (curseries - series)
	chain = *structs.Makechain(series, port, lenservers)
	recvNum = 0
	sendNum = 0

	m := "Head server: " + chain.Head + ", Tail server:" + chain.Tail
	utils.LogCEvent(chain.Server, "Client started!"+m)
	conn := createUDPSocket("127.0.0.1:" + os.Args[1])
	clientIdx := 999 - port%1000
	waitTime := time.Duration(utils.GetConfigInt("requestTimeout"))

	//wait for servers to start up
	time.Sleep(2000 * time.Millisecond)
	go simulate(conn, port, clientIdx, waitTime)
	http.HandleFunc("/alterChain", alterChainHandler)
	err := http.ListenAndServe("127.0.0.1:"+os.Args[1], nil)
	if err != nil {
		log.Fatal(err)
	}
}
