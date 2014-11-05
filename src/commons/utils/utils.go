package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	gson "github.com/bitly/go-simplejson"
)

func NewID() string {
	uuid := make([]byte, 16)
	io.ReadFull(rand.Reader, uuid)
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x", uuid[0:4])
}

func GetCount() string {
	id, err := ioutil.ReadFile("/Users/ram/deistests/src/github.com/replicasystem/src/server/counter")
	if err != nil {
		return "0"
	}
	return strings.TrimSpace(string(id))
}

func PutCount(version string) error {
	err := ioutil.WriteFile("/Users/ram/deistests/src/github.com/replicasystem/src/server/counter", []byte(version), 0644)
	if err != nil {
		return err
	}
	return nil
}

func Logoutput(server, reqid, outcome string, balance int, trans string) {
	var name string
	if server == "client" {
		name = "/Users/ram/deistests/src/github.com/replicasystem/logs/clogs"
	} else {
		name = "/Users/ram/deistests/src/github.com/replicasystem/logs/slogs"
	}
	f, _ := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)
	log.Printf("%s :%s-%s-%d-%s", server, reqid, outcome, balance, trans)
}

func Logevent(server, reqid, event string) {
	var name string
	if server == "client" {
		name = "/Users/ram/deistests/src/github.com/replicasystem/logs/clogs"
	} else {
		name = "/Users/ram/deistests/src/github.com/replicasystem/logs/slogs"
	}
	f, _ := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)
	log.Printf("%s :%s-%s", server, reqid, event)
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
	js, _ := gson.NewJson(GetFileBytes("/Users/ram/deistests/src/github.com/replicasystem/config/request.json"))
	command, _ := js.Get(data).String()
	return command
}

func Getconfig(data string) string {
	js, _ := gson.NewJson(GetFileBytes("/Users/ram/deistests/src/github.com/replicasystem/config/config.json"))
	command, _ := js.Get(data).String()
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
