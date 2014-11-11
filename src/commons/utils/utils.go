package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gson "github.com/bitly/go-simplejson"
)

var config, request *gson.Json

func SetConfigFile(filename string) {
	configFile := GetWorkDir() + "config/" + filename
	requestFile := GetWorkDir() + "config/request.json"
	config, _ = gson.NewJson(GetFileBytes(configFile))
	request, _ = gson.NewJson(GetFileBytes(requestFile))
}

func NewID() string {
	uuid := make([]byte, 16)
	io.ReadFull(rand.Reader, uuid)
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x", uuid[0:4])
}

func GetCount() string {
	id, err := ioutil.ReadFile(GetWorkDir() + "src/server/counter")
	if err != nil {
		return "0"
	}
	return strings.TrimSpace(string(id))
}

func PutCount(version string) error {
	err := ioutil.WriteFile(GetWorkDir()+"src/server/counter", []byte(version), 0644)
	if err != nil {
		return err
	}
	return nil
}

func Logoutput(server, servType, reqid, account, outcome string, balance int, trans string) {
	var name string
	if servType == "client" {
		name = GetWorkDir() + "logs/clogs"
	} else {
		name = GetWorkDir() + "logs/slogs"
	}
	f, _ := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)
	s := fmt.Sprintf("%s, requestID %s, %s, Balance: %d, %s", server, reqid, outcome, balance, trans)
	fmt.Println(s)
	log.Printf(s)
}

func LogServer(server, reqID, account, outcome, trans string, balance int) {
	Logoutput(server, "server", reqID, account, outcome, balance, trans)
}

func LogClient(server, reqID, account, outcome, trans string, balance int) {
	Logoutput(server, "client", reqID, account, outcome, balance, trans)
}

func Logevent(server, servType, event string) {
	var name string
	if servType == "client" {
		name = GetWorkDir() + "logs/clogs"
	} else if servType == "server" {
		name = GetWorkDir() + "logs/slogs"
	} else {
		name = GetWorkDir() + "logs/mlogs"
	}

	f, _ := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	fmt.Printf("%s EVENT: %s\n", server, event)
	log.SetOutput(f)
	log.Printf("%s EVENT: %s", server, event)
}

func LogSEvent(server, event string) {
	Logevent(server, "server", event)
}

func LogCEvent(server, event string) {
	Logevent(server, "client", event)
}

func LogMEvent(server, event string) {
	Logevent(server, "master", event)
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
	log.Printf("%s %s:#%d %s", server, msgType, num, msg)
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

func Getvalue(data string) string {
	command, _ := request.Get(data).String()
	return command
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

func Timeout(msg string, seconds time.Duration, f func()) error {
	c := make(chan bool)
	go func() {
		time.Sleep(seconds)
		c <- true
	}()
	go func() {
		f()
		c <- false
	}()
	if <-c && msg != "" {
		return errors.New(msg + "timed out")
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
