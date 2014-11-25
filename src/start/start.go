package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/replicasystem/src/commons/utils"
)

func exe_cmd(cmd string, wg *sync.WaitGroup) {
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0], parts[1], parts[2]).Output()
	if err != nil {
		fmt.Println("error occured when starting processes")
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
	wg.Done()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run start/start.go <config file>")
		return
	}
	logPath := os.Getenv("GOPATH") + "/src/github.com/replicasystem/logs/"
	os.RemoveAll(logPath)
	os.Mkdir(logPath, 0777)

	utils.SetConfigFile(os.Args[1])
	totalchains, _ := strconv.Atoi(utils.Getconfig("chains"))
	series, _ := strconv.Atoi(utils.Getconfig("chain1series"))
	lenservers, _ := strconv.Atoi(utils.Getconfig("chainlength"))
	clientNum := utils.GetConfigInt("clientNum")

	fmt.Println(totalchains)
	wg := new(sync.WaitGroup)
	commands := make([]string, 0, 1)
	for i := 0; i < totalchains; i++ {
		//start servers
		curSeries := 1000 * (series + i)
		for start := curSeries + 1; start <= curSeries+lenservers; start++ {
			strin := utils.GetBinDir() + "server " + strconv.Itoa(start) + " " + os.Args[1]
			commands = append(commands, strin)
		}
	}

	master := utils.GetBinDir() + "master " + os.Args[1] + " 1"
	commands = append(commands, master)

	for i := 0; i < totalchains; i++ {
		curSeries := 1000 * (series + i)
		//start clients
		for start := curSeries + 999; start > curSeries+999-clientNum; start-- {

			client := utils.GetBinDir() + "client " + strconv.Itoa(start) + " " + os.Args[1]
			commands = append(commands, client)
		}
	}
	// commands := []string{"/Users/ram/deistests/bin/server 4001", "/Users/ram/deistests/bin/server 4002", "/Users/ram/deistests/bin/server 4003"}
	fmt.Println(len(commands))
	for _, str := range commands {
		fmt.Println(str)
		wg.Add(1)
		go exe_cmd(str, wg)
	}
	wg.Wait()
}
