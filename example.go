package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

func main() {

	// FIXME: sleep a bit before curling
	time.Sleep(2000 * time.Millisecond)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("not reachable:\n%v", err)
	}
	body, err := ioutil.ReadAll(response.Body)
	//fmt.Println()\w[-._\w]*
	r, err := regexp.Compile(`<a href="/\w[/\w.]*"`)
	res := r.FindAllString(string(body), -1)
	return res

}
