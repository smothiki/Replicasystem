package main

import (
	"encoding/json"
	"fmt"
	"github.com/replicasystem/src/commons/structs"
	"github.com/replicasystem/src/commons/utils"
	"log"
	"net"
	"time"
)

const MAXLINE = 1024
const CHECK_CYCLE = 5000

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

func readOnlineMsg(conn *net.UDPConn, statMap *map[int]int) {
	for {
		buf := make([]byte, MAXLINE)
		n, sourceAddr, err := conn.ReadFromUDP(buf)

		if err != nil {
			log.Fatal(err)
		}

		var r int
		json.Unmarshal(buf[:n], &r)
		if r == 1 {
			key := sourceAddr.Port - 100
			if _, exists := (*statMap)[key]; exists {
				(*statMap)[key]++
			} else {
				//extend
			}
		}
	}
}

func checkStatus(statMap *map[int]int) {
	for {
		for serverIdx, msgRecvd := range *statMap {
			if msgRecvd == 0 {
				//failure
			}
			(*statMap)[serverIdx] = 0
			fmt.Println(serverIdx, msgRecvd)
		}
		time.Sleep(CHECK_CYCLE * time.Millisecond)
	}
}

func main() {
	utils.SetConfigFile("config.json")
	chainNum := utils.GetConfigInt("chains")
	chain1Series := utils.GetConfigInt("chian1series")
	chainLen := utils.GetConfigInt("chainlength")
	//key: port number, value : msgs received within timeframe
	servStatus := make(map[int]int)

	//init
	for i := chain1Series; i < chain1Series+chainNum; i++ {
		//for each chain
		for j := i*1000 + 1; j <= i*1000+chainLen; j++ {
			//for each server
			servStatus[j] = 0
		}
	}

	conn := createUDPSocket()
	go readOnlineMsg(conn, &servStatus)
	checkStatus(&servStatus)
	//fmt.Println(servStatus)
}
