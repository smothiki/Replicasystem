package structs

import (
	"fmt"
	"strconv"

	gson "github.com/bitly/go-simplejson"
	"github.com/replicasystem/src/commons/utils"
)

func gettypeList(prob int, typet string) *[]Request {
	listreqs := make([]Request, 0, 1)
	js, _ := gson.NewJson(utils.GetFileBytes(utils.GetWorkDir() + "config/request01.json"))
	getReqs := js.Get("requests").Get(typet)
	a, _ := getReqs.Array()
	fmt.Println(len(a))
	fmt.Println(prob)
	if len(a)-prob < 0 {
		for i := 0; i < prob-len(a); i++ {
			listreqs = append(listreqs, *Genrequest(0, "getbalance"))
		}
		for i := 0; i < len(a); i++ {
			reqid, _ := getReqs.GetIndex(i).Get("requestid").String()
			account, _ := getReqs.GetIndex(i).Get("account").String()
			amounts, _ := getReqs.GetIndex(i).Get("amount").String()
			amount, _ := strconv.Atoi(amounts)
			typet, _ := getReqs.GetIndex(i).Get("transaction").String()
			outcome, _ := getReqs.GetIndex(i).Get("outcome").String()
			listreqs = append(listreqs, *Makereply(reqid, account, outcome, typet, amount, 0))
		}
	} else {
		for i := 0; i < prob; i++ {
			reqid, _ := getReqs.GetIndex(i).Get("requestid").String()
			account, _ := getReqs.GetIndex(i).Get("account").String()
			amounts, _ := getReqs.GetIndex(i).Get("amount").String()
			amount, _ := strconv.Atoi(amounts)
			typet, _ := getReqs.GetIndex(i).Get("transaction").String()
			outcome, _ := getReqs.GetIndex(i).Get("outcome").String()
			listreqs = append(listreqs, *Makereply(reqid, account, outcome, typet, amount, 0))
		}
	}
	return &listreqs
}

func GetrequestList(prob int, typet string) *[]Request {
	listreqs := make([]Request, 0, 1)
	totalreqs, _ := strconv.Atoi(utils.Getconfig("MaxRequests"))
	types := []string{"getbalance", "deposit", "withdraw"}
	if prob == 0 {
		for i := 0; i < 3; i++ {
			for _, request := range *gettypeList(6, types[i]) {
				listreqs = append(listreqs, request)
			}
		}
		listreqs = append(listreqs, *Genrequest(0, "getbalance"))
		listreqs = append(listreqs, *Genrequest(0, "deposit"))
	} else {
		rem := totalreqs - (2 * prob)
		rem = rem / 2
		for _, request := range *gettypeList(prob*2, typet) {
			listreqs = append(listreqs, request)
		}
		for i := 0; i < 3; i++ {
			if typet != types[i] {
				for _, request := range *gettypeList(rem, types[i]) {
					listreqs = append(listreqs, request)
				}
			}
		}
	}
	for _, request := range listreqs {
		fmt.Println(request)
	}
	return &listreqs
}

func GetTestreqs() *[]Request {
	listreqs := make([]Request, 0, 1)
	js, _ := gson.NewJson(utils.GetFileBytes(utils.GetWorkDir() + "config/request01.json"))
	getReqs := js.Get("requests").Get("tests")
	a, _ := getReqs.Array()
	for i := 0; i < len(a); i++ {
		reqid, _ := getReqs.GetIndex(i).Get("requestid").String()
		account, _ := getReqs.GetIndex(i).Get("account").String()
		amounts, _ := getReqs.GetIndex(i).Get("balance").String()
		amount, _ := strconv.Atoi(amounts)
		typet, _ := getReqs.GetIndex(i).Get("transaction").String()
		outcome, _ := getReqs.GetIndex(i).Get("outcome").String()
		listreqs = append(listreqs, *Makereply(reqid, account, outcome, typet, amount, 0))
	}
	return &listreqs
}
