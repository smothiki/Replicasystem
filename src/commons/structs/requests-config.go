package structs

import (
	"math/rand"
	"strconv"

	gson "github.com/bitly/go-simplejson"
	"github.com/replicasystem/src/commons/utils"
)

func GenRequestList(rqstFile string) *[]Request {
	listreqs := make([]Request, 0, 1)
	js, _ := gson.NewJson(utils.GetFileBytes(utils.GetWorkDir() + "config/" + rqstFile))
	getReqs := js.Get("requests").Get("tests")
	arr, _ := getReqs.Array()
	numReqTest := len(arr)
	maxNumReq := utils.GetConfigInt("MaxRequests")
	if numReqTest > maxNumReq {
		numReqTest = maxNumReq
	}

	//get testcases in request file
	for i := 0; i < numReqTest; i++ {
		reqid, _ := getReqs.GetIndex(i).Get("requestid").String()
		account, _ := getReqs.GetIndex(i).Get("account").String()
		amounts, _ := getReqs.GetIndex(i).Get("amount").String()
		amount64, _ := strconv.ParseFloat(amounts, 32)
		amount := float32(amount64)
		typet, _ := getReqs.GetIndex(i).Get("transaction").String()
		outcome, _ := getReqs.GetIndex(i).Get("outcome").String()
		listreqs = append(listreqs, *Makereply(reqid, account, outcome, typet, amount, 0))
	}

	//fill rest vacancies with random requests
	if numReqTest < maxNumReq {
		js, _ := gson.NewJson(utils.GetFileBytes(utils.GetWorkDir() + "config/" + rqstFile))
		getbalanceProb := getRequestProb(js, "getbalance")
		depositProb := getRequestProb(js, "deposit")
		//withdrawProb := getRequestProb(js, "withdraw")
		sseed, _ := js.Get("requests").Get("seed").String()
		seed, _ := strconv.ParseInt(sseed, 10, 64)
		rand.Seed(seed)

		remainReqNum := maxNumReq - numReqTest
		types := []string{"getbalance", "deposit", "withdraw"}
		counter := []int{0, 0, 0}
		numTypes := []int{0, 0, 0}
		numTypes[0] = int(getbalanceProb * float32(remainReqNum))
		numTypes[1] = int(depositProb * float32(remainReqNum))
		numTypes[2] = remainReqNum - numTypes[0] - numTypes[1]
		numAccounts := remainReqNum / 4
		accounts := make([]string, numAccounts)
		for i := 0; i < numAccounts; i++ {
			accounts[i] = utils.NewID()
		}

		r := rand.New(rand.NewSource(99))
		for {
			typeIdx := rand.Intn(len(types))
			accIdx := rand.Intn(numAccounts)
			if counter[typeIdx] == numTypes[typeIdx] {
				continue
			}
			id := utils.NewID()
			amount := r.Float32() * 30
			listreqs = append(listreqs, *Makereply(id, accounts[accIdx], "none", types[typeIdx], amount, 0))
			counter[typeIdx]++
			numReqTest++
			if numReqTest == maxNumReq {
				break
			}
		}
	}
	println(listreqs)
	return &listreqs
}

func getRequestProb(js *gson.Json, request string) float32 {
	strRes, _ := js.Get("requests").Get(request + "Prob").String()
	return utils.ParseFloat32(strRes)
}
