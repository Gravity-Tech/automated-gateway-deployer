package main

import (
	"flag"
	"fmt"
	"github.com/Gravity-Tech/automated-gateway-deployer/deployer"
)

var config string

func init() {
	flag.StringVar(&config, "config", "config.json", "config path")
	flag.Parse()
}

func main() {
	_, err := deployer.Deploy(config)

	if err != nil {
		fmt.Printf(err.Error())
	}
}
