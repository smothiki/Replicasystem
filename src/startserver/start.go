package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/replicasystem/src/commons/utils"
)

func exe_cmd(cmd string, wg *sync.WaitGroup) {
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0], parts[1]).Output()
	if err != nil {
		fmt.Println("error occured")
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
	wg.Done()
}

func main() {
	utils.SetConfigFile("config.json")
	totalchains, _ := strconv.Atoi(utils.Getconfig("chains"))
	series, _ := strconv.Atoi(utils.Getconfig("chian1series"))
	lenservers, _ := strconv.Atoi(utils.Getconfig("chainlength"))
	clientNum := utils.GetConfigInt("clientNum")

	fmt.Println(totalchains)
	wg := new(sync.WaitGroup)
	commands := make([]string, 0, 1)
	for i := 0; i < totalchains; i++ {
		//start servers
		curSeries := 1000 * (series + i)
		for start := curSeries + 1; start <= curSeries+lenservers; start++ {
			strin := utils.GetBinDir() + "server " + strconv.Itoa(start)
			commands = append(commands, strin)

		}

		//start clients
		for start := curSeries + 999; start > curSeries+999-clientNum; start-- {

			client := utils.GetBinDir() + "client " + strconv.Itoa(start)
			commands = append(commands, client)
		}
		series = series + 1
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
