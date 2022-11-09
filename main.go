package main

import (
	"fmt"
	"os"

	"k8s.io/component-base/cli"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"cmio-scheduler/pkg/plugins/avacpu"
	"cmio-scheduler/pkg/plugins/avamem"
	"cmio-scheduler/pkg/plugins/io"
)

func main() {
	fmt.Println("IO scheduler started!")
	command := app.NewSchedulerCommand(
		app.WithPlugin(io.Name, io.New),
		app.WithPlugin(avacpu.Name, avacpu.New),
		app.WithPlugin(avamem.Name, avamem.New),
	)
	code := cli.Run(command)
	os.Exit(code)
}
