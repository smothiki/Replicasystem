package structs

import (
	"fmt"
	"net"
	"strconv"

	"github.com/replicasystem/src/commons/utils"
)

//Request and Reply
type Request struct {
	Requestid   string
	Account     string
	Balance     float32
	Transaction string
	Outcome     string
	Amount      float32
	Client      net.UDPAddr
	Time        string
	//for transfer
	DestBank    string
	DestAccount string
	Receiver    net.UDPAddr
	Sender      net.UDPAddr
}

type Chain struct {
	Head            string
	Tail            string
	Prev            string
	Next            string
	Server          string
	Ishead          bool
	Istail          bool
	MsgCnt          int
	Online          bool
	FailOnReqSent   bool
	FailOnRecvSent  bool
	FailOnExtension bool
	Available       bool
	FailOnRecvTrans bool
	FailOnSendTrans bool
}

type ClientNotify struct {
	Head string
	Tail string
}

type Ack struct {
	ReqKey string
}

type DestHeadRqst struct {
	DestBank string
	Sender   string
}

//for master use
type ChainInfo struct {
	Head int
	Tail int
}

func (r *Request) String(strType string) string {
	switch strType {
	case "REPLY":
		return fmt.Sprintf("reqID %s, a/c %s, %s, balance %.2f, %s, reqTime %s",
			r.Requestid, r.Account, r.Transaction, r.Balance, r.Outcome, r.Time)
	case "TRANS_REPLY":
		return fmt.Sprintf("reqID %s, from a/c %s to a/c %s at Bank %s, %s, balance %.2f, %s, reqTime %s",
			r.Requestid, r.Account, r.DestAccount, r.DestBank, r.Transaction, r.Balance, r.Outcome, r.Time)
		//return fmt.Sprintf("reqID %s, a/c %s, %s, balance %.2f, %s, reqTime %s cli %s rec %s", r.Requestid, r.Account, r.Transaction, r.Balance, r.Outcome, r.Time, r.Client.String(), r.Receiver.String())
	case "REQUEST":
		return fmt.Sprintf("reqID %s, a/c %s, %s(%.2f) %s",
			r.Requestid, r.Account, r.Transaction, r.Amount, r.Time)
	case "HISTORY":
		return fmt.Sprintf("reqID %s, a/c %s, %s(%.2f), balance %.2f, reqTime %s",
			r.Requestid, r.Account, r.Transaction, r.Amount, r.Balance, r.Time)
	case "TRANS_REQ":
		return fmt.Sprintf("reqID %s from a/c %s to a/c %s at Bank %s, %s(%.2f) %s",
			r.Requestid, r.Account, r.DestAccount, r.DestBank, r.Transaction, r.Amount, r.Time)
	case "TRANS_HIST":
		return fmt.Sprintf("reqID %s , a/c %s to a/c %s at Bank %s, %s(%.2f), balance %.2f reqTime %s",
			r.Requestid, r.Account, r.DestAccount, r.DestBank, r.Transaction, r.Amount, r.Balance, r.Time)
	default:
		return ""
	}
}

func (req *Request) MakeKey() string {
	return req.Requestid + "." + req.Time
}

func (c *Chain) String() string {
	return fmt.Sprintf("Prev: %s, Next: %s, isHead: %t, isTail: %t, isOnline: %t", c.Prev, c.Next, c.Ishead, c.Istail, c.Online)
}

func (c *Chain) PrintHeadTail() string {
	return fmt.Sprintf("Head: %s, Tail :%s", c.Head, c.Tail)
}

func (c *Chain) SetChain(cc *Chain) {
	c.Head = cc.Head
	c.Tail = cc.Tail
	c.Prev = cc.Prev
	c.Next = cc.Next
	c.Server = cc.Server
	c.Ishead = cc.Ishead
	c.Istail = cc.Istail
	c.MsgCnt = cc.MsgCnt
	c.Online = cc.Online
}

//Makechain returns pointer of a chain structure constructed from parameters
func Makechain(series, server, length int) *Chain {
	chain := &Chain{
		//next two fields are used only by client
		Head:            "",
		Tail:            "",
		Prev:            "",
		Next:            "",
		Server:          "127.0.0.1:" + strconv.Itoa(server),
		Ishead:          false,
		Istail:          false,
		MsgCnt:          0,
		Online:          true,
		FailOnReqSent:   false,
		FailOnRecvSent:  false,
		FailOnExtension: false,
		Available:       true,
	}

	if utils.GetStartDelay(server%1000-1) > 0 {
		chain.Online = false
	}

	base := series * 1000
	start := 0
	end := 0
	for i := 0; i < length; i++ {
		if utils.GetStartDelay(i) == 0 {
			start = base + i + 1
			break
		}
	}

	for i := length - 1; i >= 0; i-- {
		if utils.GetStartDelay(i) == 0 {
			end = base + i + 1
			break
		}
	}
	chain.Head = "127.0.0.1:" + strconv.Itoa(start)
	chain.Tail = "127.0.0.1:" + strconv.Itoa(end)

	if chain.Head == chain.Server {
		chain.Ishead = true
	} else {
		prev := 0
		for i := server%1000 - 2; i >= 0; i-- {
			if utils.GetStartDelay(i) == 0 {
				prev = base + i + 1
				break
			}
		}
		chain.Prev = "127.0.0.1:" + strconv.Itoa(prev)
	}

	if chain.Tail == chain.Server {
		chain.Istail = true
	} else {
		next := 0
		for i := server % 1000; i < length; i++ {
			if utils.GetStartDelay(i) == 0 {
				next = base + i + 1
				break
			}
		}
		chain.Next = "127.0.0.1:" + strconv.Itoa(next)
	}
	return chain
}

func Makereply(reqid, account, outcome, typet, destAccount, destBank string, amount, balance float32) *Request {
	rep := &Request{
		Requestid:   reqid,
		Account:     account,
		Amount:      amount,
		Outcome:     outcome,
		Transaction: typet,
		Balance:     balance,
		DestBank:    destBank,
		DestAccount: destAccount,
	}
	rep.Sender = net.UDPAddr{
		Port: 0,
		IP:   net.ParseIP("0.0.0.0"),
	}
	rep.Receiver = net.UDPAddr{
		Port: 0,
		IP:   net.ParseIP("0.0.0.0"),
	}
	rep.Client = net.UDPAddr{
		Port: 0,
		IP:   net.ParseIP("0.0.0.0"),
	}
	return rep
}
