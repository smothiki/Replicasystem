package bank

import (
	"errors"
	"fmt"

	"github.com/replicasystem/src/commons/structs"
)

type transaction struct {
	tmap map[string]string
}

func inittransaction() *transaction {
	var t transaction
	t.tmap = make(map[string]string)
	return &t
}

func (t *transaction) recordtransaction(reqid, typet string) {
	t.tmap[reqid] = typet
}

func (t *transaction) checktransaction(reqid, typet string) string {
	if name, ok := t.tmap[reqid]; ok {
		if name == typet {
			return "processed"
		} else {
			return "inconsistent"
		}
	}
	return "new"
}

type account struct {
	accountid string
	balance   int
}

func (a *account) getbalance() int {

	return a.balance
}

func (a *account) deposit(amount int) {
	a.balance = a.balance + amount
}

func (a *account) withdraw(amount int) error {
	temp := a.balance - amount
	if temp < 0 {
		err := errors.New("no funds")
		return err
	}
	a.balance = temp
	return nil
}

type Bank struct {
	amap     map[string]*account
	bankname string
	bankid   string
	t        *transaction
}

func Initbank(name, Id string) *Bank {
	var b = Bank{
		amap:     make(map[string]*account),
		bankname: name,
		bankid:   Id,
		t:        inittransaction(),
	}
	return &b
}

func (b *Bank) CheckId(accountId string) {
	if _, ok := b.amap[accountId]; ok != true {
		fmt.Println("checkid")
		fmt.Println(ok)
		newaccnt := &account{
			accountid: accountId,
			balance:   0,
		}
		fmt.Println(newaccnt)
		b.amap[accountId] = newaccnt
	}
	fmt.Println("out-checkid")
	fmt.Println(b)
}

func (b *Bank) Deposit(req *structs.Request) *structs.Request {
	b.CheckId(req.Account)
	a := b.amap[req.Account]
	resp := b.t.checktransaction(req.Requestid, "deposit")
	if resp == "new" {
		resp = "processed"
		a.deposit(req.Balance)
		b.t.recordtransaction(req.Requestid, "deposit")
	}
	return structs.Makereply(req.Requestid, req.Account, resp, "deposit", a.getbalance())
}

func (b *Bank) Withdraw(req *structs.Request) *structs.Request {
	b.CheckId(req.Account)
	a := b.amap[req.Account]
	resp := b.t.checktransaction(req.Requestid, "withdraw")
	if resp == "new" {
		resp = "processed"
		b.t.recordtransaction(req.Requestid, "withdraw")
		if err := a.withdraw(req.Balance); err != nil {
			return structs.Makereply(req.Requestid, req.Account, "insufficientfunds", "withdraw", a.getbalance())
		}
	}
	return structs.Makereply(req.Requestid, req.Account, resp, "withdraw", a.getbalance())
}

func (b *Bank) Set(rep *structs.Request) {
	b.CheckId(rep.Account)
	a := b.amap[rep.Account]
	fmt.Println(rep.Account)
	fmt.Println(a)
	a.balance = rep.Balance
	b.t.recordtransaction(rep.Requestid, rep.Transaction)
}

func (b *Bank) GetBalance(req *structs.Request) *structs.Request {
	b.CheckId(req.Account)
	a := b.amap[req.Account]
	fmt.Println("getbalacne")
	fmt.Println(a)
	//b.t.recordtransaction(req.Requestid, "getbalance")
	return structs.Makereply(req.Requestid, req.Account, "processed", "getbalance", a.getbalance())
}
