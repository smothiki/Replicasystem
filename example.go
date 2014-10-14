package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
	// "io/ioutil"
	// "net/http"
	// "regexp"
	// "time"
)

type jaffa struct {
	a int
}

func main() {

	a := []jaffa{
		jaffa{a: 1},
		jaffa{a: 2},
		jaffa{a: 3},
	}

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
