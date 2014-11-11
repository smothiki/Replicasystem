package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/replicasystem/src/commons/structs"
	"github.com/replicasystem/src/commons/utils"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const MAXLINE = 1024
const CHECK_CYCLE = 3000

var recvNum, sendNum int
var master string

func logEvent(event string) {
	utils.LogMEvent(master, event)
}

func logMsg(msgType, msg string) {
	if msgType == "SENT" {
		utils.LogMMsg("", msgType, sendNum, msg)
		sendNum++
	} else if msgType == "RECV" {
		utils.LogMMsg("", msgType, sendNum, msg)
		recvNum++
	} else {
		log.Println("LOG ERROR: UNKOWN MSG TYPE")
	}
}

func createUDPSocket() *net.UDPConn {
	s := utils.Getconfig("master")
	ip, port := structs.GetIPAndPort(s)
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

func readOnlineMsg(conn *net.UDPConn, statMap *map[string]*structs.Chain) {
	for {
		buf := make([]byte, MAXLINE)
		n, sourceAddr, err := conn.ReadFromUDP(buf)

		if err != nil {
			log.Fatal(err)
		}

		var r int
		json.Unmarshal(buf[:n], &r)
		if r == 1 {
			keyPort := strconv.Itoa(sourceAddr.Port - 100)
			key := "127.0.0.1:" + keyPort
			(*statMap)[key].MsgCnt++
			logMsg("RECV", "ONLINE (from "+key+")")
		}
	}
}

func checkStatus(statMap *map[string]*structs.Chain) {
	for {
		time.Sleep(CHECK_CYCLE * time.Millisecond)
		for serverIdx, chain := range *statMap {
			if chain.MsgCnt == 0 && chain.Online {
				//failure
				logEvent("server" + serverIdx + " failed")
				alterChain(serverIdx, statMap)
			} else if chain.MsgCnt > 0 && !chain.Online {
				//extend
				logEvent("new server" + serverIdx + "online")
				extendChain(serverIdx, statMap)
			}
			//fmt.Println(serverIdx, chain)
			(*statMap)[serverIdx].MsgCnt = 0
		}
	}
}

func extendChain(newTail string, statMap *map[string]*structs.Chain) {
	// find tail
	var oldTail string
	for serverIdx, chain := range *statMap {
		if chain.Online && chain.Istail {
			oldTail = serverIdx
			break
		}
	}

	fmt.Println("oldTail", oldTail)
	fmt.Println("newTail", newTail)

	// notify new Tail
	(*statMap)[newTail].Istail = true
	(*statMap)[newTail].Prev = oldTail
	(*statMap)[newTail].Next = ""
	(*statMap)[newTail].Online = true
	notifyServer(newTail, "extendChain", (*statMap)[newTail])

	// notify old tail
	(*statMap)[oldTail].Istail = false
	(*statMap)[oldTail].Next = newTail
	notifyServer(oldTail, "extendChain", (*statMap)[oldTail])

	// notify clients
	newHeadTail := structs.ClientNotify{
		Head: "",
		Tail: newTail,
	}
	notifyClients(&newHeadTail)
}

func notifyServer(dest, action string, newChain *structs.Chain) {
	msg, _ := json.Marshal(newChain)
	client := &http.Client{}
	req, _ := http.NewRequest("POST",
		"http://"+dest+"/"+action, bytes.NewBuffer(msg))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	_, err := client.Do(req)
	if err != nil {
		log.Println("ERROR whiile notifying server chain modification", err)
	}
	logMsg("SENT", newChain.String())
}

func notifyClient(dest string, data *structs.ClientNotify) {
	msg, _ := json.Marshal(data)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://"+dest+"/alterChain", bytes.NewBuffer(msg))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	_, err := client.Do(req)
	if err != nil {
		log.Println("ERROR while notifying client chain modification", err)
	}
	logMsg("SENT", string(msg))
}

func notifyClients(data *structs.ClientNotify) {
	var serverPort int
	if data.Head != "" {
		_, serverPort = structs.GetIPAndPort(data.Head)
	} else {
		_, serverPort = structs.GetIPAndPort(data.Tail)
	}
	clientPortStart := serverPort - serverPort%1000 + 999
	numClient := utils.GetConfigInt("clientNum")
	for i := clientPortStart; i > clientPortStart-numClient; i-- {
		dest := "127.0.0.1:" + strconv.Itoa(i)
		notifyClient(dest, data)
	}
}

func alterChain(server string, statMap *map[string]*structs.Chain) {
	curNode := (*statMap)[server]
	nextKey := (*statMap)[server].Next
	prevKey := (*statMap)[server].Prev
	(*statMap)[server].Online = false
	if curNode.Ishead && !curNode.Istail {
		(*statMap)[nextKey].Prev = ""
		(*statMap)[nextKey].Ishead = true
		notifyServer(nextKey, "alterChain", (*statMap)[nextKey])
		newHeadTail := structs.ClientNotify{
			Head: nextKey,
			Tail: "",
		}
		logEvent("Head becomes " + nextKey)
		notifyClients(&newHeadTail)
	} else if !curNode.Ishead && curNode.Istail {
		(*statMap)[prevKey].Next = ""
		(*statMap)[prevKey].Istail = true
		notifyServer(prevKey, "alterChain", (*statMap)[prevKey])
		newHeadTail := structs.ClientNotify{
			Head: "",
			Tail: prevKey,
		}
		logEvent("Head becomes " + prevKey)
		notifyClients(&newHeadTail)
	} else if !curNode.Ishead && !curNode.Istail {
		(*statMap)[prevKey].Next = curNode.Next
		(*statMap)[nextKey].Prev = curNode.Prev
		notifyServer(nextKey, "alterChain", (*statMap)[nextKey])
		notifyServer(prevKey, "alterChain", (*statMap)[prevKey])
	} else {
		fmt.Println("ERROR: no server available")
	}
}

func main() {
	log.Println("master started!")
	utils.SetConfigFile(os.Args[1])
	master = "127.0.0.1:65535"
	chainNum := utils.GetConfigInt("chains")
	chain1Series := utils.GetConfigInt("chian1series")
	chainLen := utils.GetConfigInt("chainlength")
	//key: port number, value : msgs received within timeframe
	servStatus := make(map[string]*structs.Chain)

	//init
	for i := chain1Series; i < chain1Series+chainNum; i++ {
		//for each chain
		for j := i*1000 + 1; j <= i*1000+chainLen; j++ {
			//for each server
			key := "127.0.0.1:" + strconv.Itoa(j)
			servStatus[key] = structs.Makechain(i, j, chainLen)
		}
	}

	conn := createUDPSocket()
	go readOnlineMsg(conn, &servStatus)
	checkStatus(&servStatus)
	//fmt.Println(servStatus)
}
