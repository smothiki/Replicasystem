package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"sync"

	"github.com/replicasystem/src/commons/utils"
)

func exe_cmd(cmd string, wg *sync.WaitGroup) {
	out, err := exec.Command(cmd).Output()
	if err != nil {
		fmt.Println("error occured")
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
	wg.Done()
}

func main() {
	totalchains, _ := strconv.Atoi(utils.Getconfig("chains"))
	// series, _ := strconv.Atoi(utils.Getconfig("clientseries"))

	fmt.Println(totalchains)
	wg := new(sync.WaitGroup)
	// commands := make([]string, 0, 1)
	// 	for start := 10000*series + 1; start <= 1000*series+lenservers; start++ {
	// 		strin :=
	// 		commands = append(commands, strin)
	// 	}
	// 	series = series + 1
	// }
	// commands := []string{"/Users/ram/deistests/bin/server 4001", "/Users/ram/deistests/bin/server 4002", "/Users/ram/deistests/bin/server 4003"}
	// fmt.Println(len(commands))
	// for _, str := range commands {
	// 	fmt.Println(str)
	wg.Add(1)
	go exe_cmd(utils.GetBinDir() + "client", wg)
	wg.Wait()
}
