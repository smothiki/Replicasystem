package structs

import (
	//	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/replicasystem/src/commons/utils"
)

type Request struct {
	Requestid   string
	Account     string
	Balance     int
	Transaction string
	Outcome     string
	Client      net.UDPAddr
	Time        string
}

type Chain struct {
	Head   string
	Tail   string
	Prev   string
	Next   string
	Server string
	Ishead bool
	Istail bool
	MsgCnt int
	Online bool
}

type ClientNotify struct {
	Head string
	Tail string
}

type Ack struct {
	ReqKey string
}

func (req *Request) MakeKey() string {
	return req.Requestid + "|" + req.Time
}

func Makechain(series, server, length int) *Chain {
	start := series*1000 + 1
	//fmt.Println(server)
	//fmt.Println(start)
	//fmt.Println(start + length - 1)
	chain := &Chain{
		//next two fields are used only by client
		Head:   "127.0.0.1:" + strconv.Itoa(start),
		Tail:   "127.0.0.1:" + strconv.Itoa(start+length-1),
		Prev:   "127.0.0.1:" + strconv.Itoa(server-1),
		Next:   "127.0.0.1:" + strconv.Itoa(server+1),
		Server: "127.0.0.1:" + strconv.Itoa(server),
		Ishead: false,
		Istail: false,
		MsgCnt: 0,
		Online: true,
	}
	if server == start {
		chain.Ishead = true
		chain.Prev = ""
	}
	if server == start+length-1 {
		chain.Istail = true
		chain.Next = ""
	}
	return chain
}

func Genrequest(balance int, typet string) *Request {
	req := &Request{
		Balance:     balance,
		Requestid:   utils.NewID(),
		Account:     utils.NewID(),
		Transaction: typet,
		Outcome:     "none"}
	return req
}

func Makereply(reqid, account, outcome, typet string, balance int) *Request {
	rep := &Request{
		Requestid:   reqid,
		Account:     account,
		Outcome:     outcome,
		Transaction: typet,
		Balance:     balance}
	return rep
}

func GetIPAndPort(server string) (string, int) {
	r := strings.Split(server, ":")
	ip := r[0]
	port, _ := strconv.Atoi(r[1])
	return ip, port
}
