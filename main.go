package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/c-bata/github-actions-ingress-rules-editor/internal/ingress"
)

// See https://developer.github.com/actions/creating-github-actions/accessing-the-runtime-environment/#exit-codes-and-statuses
const (
	exitSuccess         = 0
	exitInvalidCmdArgs  = 1
	exitK8sOperationErr = 2
	exitNeutral         = 78
)

var (
	ingressName string
	ingressHost string
	serviceName string
	servicePort int
	namespace   string
	pathRule    string
)

func run(operation string) int {
	editor, err := ingress.New(ingressName, &ingress.Option{
		Namespace: namespace,
	})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		return exitK8sOperationErr
	}

	switch operation {
	case "add":
		var added bool
		added, err = editor.Add(ingressHost, pathRule, serviceName, servicePort)
		if !added {
			return exitNeutral
		}
	case "remove":
		var found bool
		found, err = editor.Remove(ingressHost)
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

func validateCmdArgs(op string) error {
	switch op {
	case "add":
		if ingressName == "" || ingressHost == "" ||
			serviceName == "" || servicePort == -1 {
			return errors.New("must specify 'ingress', 'host', 'service' and 'port' option")
		}
	case "remove":
		if ingressName == "" || ingressHost == "" {
			return errors.New("must specify 'ingress' and 'host'")
		}
	default:
		return errors.New("command not found")
	}
	return nil
}

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), `Usage:
  %s add -ingress=<INGRESS_NAME> -host=<INGRESS_HOST> -service=<SERVICE_NAME> -port=<SERVICE_PORT>
  %s remove -ingress=<INGRESS_NAME> -host=<INGRESS_HOST>

`, os.Args[0], os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.StringVar(&ingressName, "ingress", "", "name of kubernetes ingress (required).")
	flag.StringVar(&ingressHost, "host", "", "ingress host (required).")
	flag.StringVar(&pathRule, "path", "/*", "matching path rules of an incoming request (optional).")
	flag.StringVar(&serviceName, "service", "", "name of kubernetes service (required when running 'add').")
	flag.IntVar(&servicePort, "port", -1, "port number (required when running 'add').")
	flag.StringVar(&namespace, "namespace", "", "kubernetes namespace (optional).")

	flag.Usage = usage
	flag.Parse()

	op := flag.Arg(0)
	err := validateCmdArgs(op)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(exitInvalidCmdArgs)
	}

	status := run(op)
	os.Exit(status)
}
