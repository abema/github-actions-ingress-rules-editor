package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"./internal/ingress"
)

// See https://developer.github.com/actions/creating-github-actions/accessing-the-runtime-environment/#exit-codes-and-statuses
const (
	exitSuccess         = 0
	exitInvalidCmdArgs  = 1
	exitK8sOperationErr = 2
	exitNeutral         = 78
)

var (
	namespace string
)

func validateCmdArgs(args []string) (ok bool) {
	if len(args) < 2 {
		return false
	}

	op := args[0]
	switch op {
	case "add":
		if len(args) != 5 {
			return false
		}
	case "remove":
		if len(args) != 3 {
			return false
		}
	default:
		return false
	}
	return true
}

func run(args []string) int {
	ok := validateCmdArgs(args)
	if !ok {
		return exitInvalidCmdArgs
	}

	operation := args[0]
	ingressName := args[1]
	editor, err := ingress.New(ingressName, &ingress.Option{})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		return exitK8sOperationErr
	}

	switch operation {
	case "add":
		host := args[2]
		serviceName := args[3]
		servicePort, err := strconv.Atoi(args[4])
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "error:", err)
			return exitInvalidCmdArgs
		}
		var added bool
		added, err = editor.Add(host, serviceName, servicePort)
		if !added {
			return exitNeutral
		}
	case "remove":
		host := args[2]
		var found bool
		found, err = editor.Remove(host)
		if !found {
			return exitNeutral
		}
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		return exitK8sOperationErr
	}
	return exitSuccess
}

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), `Usage:
  %s add <INGRESS_NAME> <INGRESS_HOST> <SERVICE_NAME> <SERVICE_PORT>
  %s remove <INGRESS_NAME> <INGRESS_HOST>

`, os.Args[0], os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.StringVar(&namespace, "namespace", "default", "kubernetes namespace")
	flag.Usage = usage
	flag.Parse()

	status := run(flag.Args())
	if status == exitInvalidCmdArgs {
		usage()
	}
	os.Exit(status)
}
