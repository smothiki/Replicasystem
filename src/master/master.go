package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/replicasystem/src/commons/structs"
	"github.com/replicasystem/src/commons/utils"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const MAXLINE = 1024 //max length of char buffer

var checkCycle time.Duration //interval of checking servers' status
var recvNum, sendNum int     //msg counter for logging
var master string            //master server information

//logEvent logs event to log file
func logEvent(event string) {
	utils.LogMEvent("", event)
}

//logMsg logs received / sent msg to log file, msgType is either
//"SENT" or "RECV", counterServer is corresponding sender / receiver
func logMsg(msgType, msg, counterServer string) {
	if msgType == "SENT" {
		msg += " (to " + counterServer + ")"
		utils.LogMMsg("", msgType, sendNum, msg)
		sendNum++
	} else if msgType == "RECV" {
		msg += " (from " + counterServer + ")"
		utils.LogMMsg("", msgType, recvNum, msg)
		recvNum++
	} else {
		log.Println("LOG ERROR: UNKOWN MSG TYPE")
	}
}

//createUDPSocket creates and listens UDP socket, through which
//servers send online (health) message to master
func createUDPSocket() *net.UDPConn {
	s := utils.Getconfig("master")
	ip, port := utils.GetIPAndPort(s)
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

//readOnlineMsg reads online (health) message from conn and sets
//msg counter for each server in statMap
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
			if !(*statMap)[key].Available {
				continue
			}
			(*statMap)[key].MsgCnt++
			logMsg("RECV", "ONLINE", key)
		}
	}
}

//checkStatus checks message counter in statMap every checkCycle
//milliseconds and sends alterChain or extendChain requests to
//servers based on message counts
func checkStatus(statMap *map[string]*structs.Chain) {
	for {
		time.Sleep(checkCycle * time.Millisecond)
		for serverIdx, chain := range *statMap {
			if !chain.Available {
				continue
			}
			if chain.MsgCnt == 0 && chain.Online {
				//failure
				logEvent("server " + serverIdx + " failed")
				alterChain(serverIdx, statMap)
			} else if chain.MsgCnt > 0 && !chain.Online {
				//extend
				if extendChain(serverIdx, statMap) {
					logEvent("new server " + serverIdx + " online")
				}
			}
			//fmt.Println(serverIdx, chain)
			(*statMap)[serverIdx].MsgCnt = 0
		}
	}
}

//extendChain notify clients and servers the change of chain,
//sets server status map statMap to new status. It returns true if no errors,
//returns false if new server fails
func extendChain(newTail string, statMap *map[string]*structs.Chain) bool {
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
	r := notifyServer(newTail, "extendChain", (*statMap)[newTail])
	if !r {
		logEvent("Server " + newTail + " failed during extension. Stop extending chain.")
		fmt.Println(newTail + " failed during extension")
		(*statMap)[newTail].Available = false
		return false
	}

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
	return true
}

//notifyServer notifies dest change of chain by sending newChain
//structure, action can be either "alterChain" or "extendChain"
//Returns true if no errors, returns false if new server fails
//when extending chains
func notifyServer(dest, action string, newChain *structs.Chain) bool {
	msg, _ := json.Marshal(newChain)
	client := &http.Client{}
	req, _ := http.NewRequest("POST",
		"http://"+dest+"/"+action, bytes.NewBuffer(msg))
	req.Header = http.Header{
		"accept": {"application/json"},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR whiile notifying server chain modification", err)
	}

	if resp == nil {
		return true
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	logMsg("SENT", newChain.String(), dest)

	if action == "extendChain" && string(body) == "failed" {
		//handle failure
		return false
	}
	return true
}

//notifyClient notifies client dest new head and tail by sending
//data structure
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
	if data.Head != "" {
		logMsg("SENT", "New head is "+data.Head, dest)
	} else {
		logMsg("SENT", "New tail is "+data.Tail, dest)
	}
}

//notifyClients notifies all clients of new head and tail
//by calling notifyClient() with data structure
func notifyClients(data *structs.ClientNotify) {
	var serverPort int
	if data.Head != "" {
		_, serverPort = utils.GetIPAndPort(data.Head)
	} else {
		_, serverPort = utils.GetIPAndPort(data.Tail)
	}
	clientPortStart := serverPort - serverPort%1000 + 999
	numClient := utils.GetConfigInt("clientNum")
	for i := clientPortStart; i > clientPortStart-numClient; i-- {
		dest := "127.0.0.1:" + strconv.Itoa(i)
		notifyClient(dest, data)
	}
}

//set server status map statMap according to change of chain.
//It's called when server fails, and the function notifies
//servers and clients the change
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
	chain1Series := utils.GetConfigInt("chain1series")
	chainLen := utils.GetConfigInt("chainlength")
	checkCycle = time.Duration(utils.GetConfigInt("checkOnlineCycle"))
	//key: server addr, value : msgs received within timeframe
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
}
