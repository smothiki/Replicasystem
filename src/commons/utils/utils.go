package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gson "github.com/bitly/go-simplejson"
)

var config *gson.Json

func SetConfigFile(filename string) {
	configFile := GetWorkDir() + "config/" + filename
	config, _ = gson.NewJson(GetFileBytes(configFile))
}

func NewID() string {
	uuid := make([]byte, 16)
	io.ReadFull(rand.Reader, uuid)
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x", uuid[0:4])
}

func LogEventData(server, servType, msgType, msg string) {
	name := GetWorkDir() + "logs/"
	switch servType {
	case "client":
		name += "clogs"
	case "server":
		name += "slogs"
	case "master":
		name += "mlogs"
	default:
		log.Fatal("ERROR while logging events/data, wrong msgType")
	}

	f, _ := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	//fmt.Printf("%s %s: %s\n", server, msgType, msg)
	log.SetOutput(f)
	log.Printf("%s %s: %s", server, msgType, msg)
}

func LogSEvent(server, event string) {
	LogEventData(server, "server", "EVENT", event)
}

func LogCEvent(server, event string) {
	LogEventData(server, "client", "EVENT", event)
}

func LogMEvent(server, event string) {
	LogEventData(server, "master", "EVENT", event)
}
func LogMsg(server, servType, msgType, msg string, num int) {
	var name string
	if servType == "client" {
		name = GetWorkDir() + "logs/clogs"
	} else if servType == "server" {
		name = GetWorkDir() + "logs/slogs"
	} else if servType == "master" {
		name = GetWorkDir() + "logs/mlogs"
	}

	f, _ := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	//log.Printf("%s %s:#%d %s", server, msgType, num, msg)
	log.SetOutput(f)
	log.Printf("%s %s:#%d %s", server, msgType, num, msg)
}

func LogSMsg(server, msgType string, num int, msg string) {
	LogMsg(server, "server", msgType, msg, num)
}

func LogCMsg(client, msgType string, num int, msg string) {
	LogMsg(client, "client", msgType, msg, num)
}

func LogMMsg(master, msgType string, num int, msg string) {
	LogMsg(master, "master", msgType, msg, num)
}

func GetFileBytes(filename string) []byte {
	file, _ := os.Open(filename)
	defer file.Close()
	stat, _ := file.Stat()
	bs := make([]byte, stat.Size())
	_, _ = file.Read(bs)
	return bs
}

func Getconfig(data string) string {
	command, _ := config.Get(data).String()
	return command
}

func GetConfigInt(data string) int {
	r, _ := strconv.Atoi(Getconfig(data))
	return r
}

func GetLifeTime(index int) int {
	command, _ := config.Get("lifetime").GetIndex(index).Int()
	return command
}

func GetStartDelay(index int) int {
	command, _ := config.Get("startDelay").GetIndex(index).Int()
	return command
}

func GetTestCaseGenMethod(index int) string {
	v, _ := config.Get("testcases").GetIndex(index).String()
	return v
}

func GetTestRequestFile(index int) string {
	v, _ := config.Get("testrequests").GetIndex(index).String()
	return v
}

func GetFailOnReqSent(index int) bool {
	return getBoolInArray(index, "failOnReqSent")
}

func GetFailOnRecvSent(index int) bool {
	return getBoolInArray(index, "failOnRecvSent")
}

func GetFailOnExtension(index int) bool {
	return getBoolInArray(index, "failOnExtension")
}

func getBoolInArray(index int, key string) bool {
	v, _ := config.Get(key).GetIndex(index).Int()
	if v == 1 {
		return true
	} else {
		return false
	}
}

func Timeout(msg string, seconds time.Duration, f func() /*, xxx int*/) error {
	tmr := time.NewTimer(seconds)
	exp := true
	go func() {
		f()
		tmr.Reset(1 * time.Millisecond)
		exp = false
	}()
	<-tmr.C

	if exp {
		//fmt.Println("EXP", xxx)
		return errors.New(msg + " timeout")
	}
	return nil
}

func SetTimer(seconds int, f func()) {
	go func() {
		time.Sleep(time.Duration(seconds*1000) * time.Millisecond)
		f()
	}()
}

func GetWorkDir() string {
	return os.Getenv("GOPATH") + "/src/github.com/replicasystem/"
}

func GetBinDir() string {
	return os.Getenv("GOBIN") + "/"
}

func ParseFloat32(n string) float32 {
	s, _ := strconv.ParseFloat(n, 32)
	return float32(s)
}

func GetIPAndPort(server string) (string, int) {
	r := strings.Split(server, ":")
	ip := r[0]
	port, _ := strconv.Atoi(r[1])
	return ip, port
}

func GetChainNum(addr string) int {
	a := strings.Split(addr, ":")
	port, _ := strconv.Atoi(a[1])
	return int(port / 1000)
}
