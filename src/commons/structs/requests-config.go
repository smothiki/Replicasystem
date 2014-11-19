package structs

import (
	"math/rand"
	"strconv"

	gson "github.com/bitly/go-simplejson"
	"github.com/replicasystem/src/commons/utils"
)

//GenRequestList generates a slice of requests. It reads parameters and
//predefined requests from rqstFile. The returned slice contains maxRequests of
//requests. If the number of predefined requests is less than maxRequests,
//remaining requests are generated randomly based on the probabilities given
//in rqstFile.
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
		//TODO: add destAccount & destBank field
		listreqs = append(listreqs, *Makereply(reqid, account, "", typet, amount, 0))
	}

	//fill rest vacancies with random requests
	if numReqTest < maxNumReq {
		js, _ := gson.NewJson(utils.GetFileBytes(utils.GetWorkDir() + "config/" + rqstFile))
		getbalanceProb := getRequestProb(js, "getbalance")
		depositProb := getRequestProb(js, "deposit")
		withdrawProb := getRequestProb(js, "withdraw")
		sseed, _ := js.Get("requests").Get("seed").String()
		seed, _ := strconv.ParseInt(sseed, 10, 64)
		rand.Seed(seed)

		remainReqNum := maxNumReq - numReqTest
		types := []string{"getbalance", "deposit", "withdraw", "transfer"}
		counter := []int{0, 0, 0}
		numTypes := []int{0, 0, 0}
		numTypes[0] = int(getbalanceProb * float32(remainReqNum))
		numTypes[1] = int(depositProb * float32(remainReqNum))
		numTypes[2] = int(withdrawProb * float32(remainReqNum))
		numTypes[3] = remainReqNum - numTypes[0] - numTypes[1] - numTypes[2]
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
			//TODO: add destAccount & destBank field randomly to random rquest list
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

//getRequestProb read probability of request from js and  convert it to float32
func getRequestProb(js *gson.Json, request string) float32 {
	strRes, _ := js.Get("requests").Get(request + "Prob").String()
	return utils.ParseFloat32(strRes)
}
