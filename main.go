package main

import (
	"github.com/j-griffith/solidfire-docker-driver/sfapi"
	"github.com/j-griffith/solidfire-docker-driver/sfcli"
	"os"
)

const (
	VERSION = "1.0.0"
)

var (
	client *sfapi.Client
)

func main() {
	cli := sfcli.NewCli(VERSION)
	cli.Run(os.Args)

}
