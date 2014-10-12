package structs

import "github.com/replicasystem/utils"

type Request struct {
	Requestid   string
	Account     string
	Balance     int
	Transaction string
}

type Reply struct {
	ReqID       string `json:"reqid"`
	AccountNum  string `json:"accountid"`
	Outcome     string `json:"outcome"`
	Balance     int    `json:"balance"`
	Transaction string `json:"Transaction"`
}

// type Chain struct {
// 	Head   string
// 	Tail   string
// 	list   []string
// 	next   string
// 	server string
// 	ishead bool
// 	istail bool
// }
//
// func Makechain(series, length int) *Chain{
// 	start := series * 1000
//
// }

func Genrequest(balance int, typet string) *Request {
	req := &Request{
		Balance:     balance,
		Requestid:   utils.NewID(),
		Account:     utils.NewID(),
		Transaction: typet}
	return req
}

func Makereply(reqid, account, outcome, typet string, balance int) *Reply {
	rep := &Reply{
		ReqID:       reqid,
		AccountNum:  account,
		Outcome:     outcome,
		Transaction: typet,
		Balance:     balance}
	return rep
}
