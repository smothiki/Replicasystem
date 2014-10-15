package structs

import (
	"fmt"
	"strconv"

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
}

func Makechain(series, server, length int) *Chain {
	start := series*1000 + 1
	fmt.Println(server)
	fmt.Println(start)
	fmt.Println(start + length - 1)
	chain := &Chain{
		Head:   "localhost:" + strconv.Itoa(start),
		Tail:   "localhost:" + strconv.Itoa(start+length),
		Next:   "localhost:" + strconv.Itoa(server+1),
		Server: "localhost:" + strconv.Itoa(server),
		Ishead: false,
		Istail: false,
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
