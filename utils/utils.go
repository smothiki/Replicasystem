package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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

func Logoutput(server, reqid, outcome string, balance int) {
	f, _ := os.OpenFile("logs", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)
	log.Printf("%s :%s-%s-%d", server, reqid, outcome, balance)
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
	js, _ := gson.NewJson(GetFileBytes("../config.json"))
	command, _ := js.Get(data).String()
	return command
}

func Timeout(msg string, seconds time.Duration, f func()) error {
	c := make(chan bool)
	// Make sure we are not too long
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
