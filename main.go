package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Gravity-Tech/automated-gateway-deployer/deployer"
	"os"
)

var config string

func init() {
	flag.StringVar(&config, "config", "config.json", "config path")
	flag.Parse()
}

func main() {
	resultConfig, err := deployer.Deploy(config)

	if err != nil {
		fmt.Printf(err.Error())
	}

	resultBytes, err := json.Marshal(&resultConfig)
	if err != nil {
		fmt.Printf(err.Error())
	}

	err = os.WriteFile("./result.json", resultBytes, 0755)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
