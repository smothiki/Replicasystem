package bank

import (
	"errors"
	"fmt"

	"github.com/replicasystem/src/commons/structs"
)

type Transaction struct {
	Tid         string
	Amount      float32
	AccountId   string
	Operation   string
	DestBank    string
	DestAccount string
}

type transactions struct {
	tmap map[string]*Transaction
}

func initTransactions() *transactions {
	var t transactions
	t.tmap = make(map[string]*Transaction)
	return &t
}

func (t *transactions) RecordTransaction(trans *Transaction) {
	t.tmap[trans.Tid] = trans
}

func (t *transactions) checkTransaction(trans *Transaction) string {
	if trans1, ok := t.tmap[trans.Tid]; ok {
		if trans1.equals(trans) {
			return "processed"
		} else {
			return "inconsistent"
		}
	}
	return "new"
}

func (t1 *Transaction) equals(t2 *Transaction) bool {
	fmt.Println("=== trans check ===")
	fmt.Println("Current opt:", t2.Tid, t2.Operation)
	fmt.Println(t1)
	fmt.Println(t2)
	fmt.Println("=== check end ===")
	return t1.Tid == t2.Tid &&
		t1.Amount == t2.Amount &&
		t1.AccountId == t2.AccountId &&
		t1.Operation == t2.Operation &&
		t1.DestBank == t2.DestBank &&
		t1.DestAccount == t2.DestAccount
}

type Account struct {
	Accountid string
	Balance   float32
}

func (a *Account) getbalance() float32 {

	return a.Balance
}

func (a *Account) deposit(amount float32) {
	a.Balance = a.Balance + amount
}

func (a *Account) withdraw(amount float32) error {
	temp := a.Balance - amount
	if temp < 0 {
		err := errors.New("no funds")
		return err
	}
	a.Balance = temp
	return nil
}

type Bank struct {
	amap     map[string]*Account
	Bankname string
	Bankid   string
	T        *transactions
}

func Initbank(name, Id string) *Bank {
	var b = Bank{
		amap:     make(map[string]*Account),
		Bankname: name,
		Bankid:   Id,
		T:        initTransactions(),
	}
	return &b
}

func (b *Bank) Accounts() *map[string]*Account {
	return &b.amap
}

func (b *Bank) TransMap() *map[string]*Transaction {
	return &b.T.tmap
}

func (b *Bank) CheckId(accountId string) {
	if _, ok := b.amap[accountId]; ok != true {
		newaccnt := &Account{
			Accountid: accountId,
			Balance:   0,
		}
		b.amap[accountId] = newaccnt
	}
}

func (b *Bank) AddAccount(id string, balance float32) {
	acc := &Account{
		Accountid: id,
		Balance:   balance,
	}
	b.amap[id] = acc
}

func MakeTransaction(r *structs.Request) *Transaction {
	t := &Transaction{
		Tid:         r.Requestid,
		Amount:      r.Amount,
		AccountId:   r.Account,
		Operation:   r.Transaction,
		DestAccount: r.DestAccount,
		DestBank:    r.DestBank,
	}
	return t
}

func (b *Bank) Deposit(req *structs.Request) *structs.Request {
	b.CheckId(req.Account)
	a := b.amap[req.Account]
	newTrans := MakeTransaction(req)
	resp := b.T.checkTransaction(newTrans)
	if resp == "new" {
		resp = "processed"
		a.deposit(req.Amount)
		b.T.RecordTransaction(newTrans)
	}
	return structs.Makereply(req.Requestid, req.Account, resp, "deposit",
		req.DestAccount, req.DestBank, req.Amount, a.getbalance())
}

func (b *Bank) Withdraw(req *structs.Request) *structs.Request {
	b.CheckId(req.Account)
	a := b.amap[req.Account]
	newTrans := MakeTransaction(req)
	resp := b.T.checkTransaction(newTrans)
	if resp == "new" {
		resp = "processed"
		b.T.RecordTransaction(newTrans)
		if err := a.withdraw(req.Amount); err != nil {
			return structs.Makereply(req.Requestid, req.Account, "insufficientfunds",
				"withdraw", req.DestAccount, req.DestBank, req.Amount, a.getbalance())
		}
	}
	return structs.Makereply(req.Requestid, req.Account, resp, "withdraw",
		req.DestAccount, req.DestBank, req.Amount, a.getbalance())
}

func (b *Bank) Set(req *structs.Request) {
	b.CheckId(req.Account)
	a := b.amap[req.Account]
	a.Balance = req.Balance
	newTrans := MakeTransaction(req)
	if req.Outcome != "inconsistent" {
		b.T.RecordTransaction(newTrans)
	}
}

func (b *Bank) GetBalance(req *structs.Request) *structs.Request {
	b.CheckId(req.Account)
	a := b.amap[req.Account]
	//b.T.recordtransaction(req.Requestid, "getbalance")
	return structs.Makereply(req.Requestid, req.Account, "processed",
		"getbalance", req.DestAccount, req.DestBank, req.Amount, a.getbalance())
}

/*bank Id identifies the current series of chain . If dest bank and bank id are same we have add the
* balance to the account . If dest bank and bank id are different we have to subtract the amount from the bank
 */

func (b *Bank) Transfer(req *structs.Request) *structs.Request {
	var account string
	if req.DestBank != b.Bankid {
		account = req.Account
	} else {
		account = req.DestAccount
	}
	b.CheckId(account)
	a := b.amap[account]
	newTrans := MakeTransaction(req)
	resp := b.T.checkTransaction(newTrans)
	if resp == "new" {
		resp = "processed"
		b.T.RecordTransaction(newTrans)
		if req.DestBank != b.Bankid {
			if (a.Balance - req.Amount) < 0 {
				return structs.Makereply(req.Requestid, req.Account, "insufficientfunds",
					"transfer", req.DestAccount, req.DestBank, req.Amount, a.getbalance())
			}
			a.Balance = a.Balance - req.Amount
		} else {
			a.Balance = a.Balance + req.Amount
		}

	}
	return structs.Makereply(req.Requestid, req.Account, resp, "transfer",
		req.DestAccount, req.DestBank, req.Amount, a.getbalance())
}
