package structs

import (
	"fmt"
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
}

type Chain struct {
	Head   string
	Tail   string
	Next   string
	Server string
	Ishead bool
	Istail bool
	Client string
}

func Makechain(series, server, length int) *Chain {
	start := series*1000 + 1
	fmt.Println(server)
	fmt.Println(start)
	fmt.Println(start + length - 1)
	chain := &Chain{
		/*
			Head:   "localhost:" + strconv.Itoa(start),
			Tail:   "localhost:" + strconv.Itoa(start+length),
			Next:   "localhost:" + strconv.Itoa(server+1),
			Server: "localhost:" + strconv.Itoa(server),
		*/
		Head:   "127.0.0.1:" + strconv.Itoa(start),
		Tail:   "127.0.0.1:" + strconv.Itoa(start+length-1),
		Next:   "127.0.0.1:" + strconv.Itoa(server+1),
		Server: "127.0.0.1:" + strconv.Itoa(server),
		Ishead: false,
		Istail: false,
		Client: "localhost:" + strconv.Itoa(series*1000),
	}
	if server == start {
		chain.Ishead = true
	}
	if server == start+length-1 {
		chain.Istail = true
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
