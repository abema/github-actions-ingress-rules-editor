package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// See https://developer.github.com/actions/creating-github-actions/accessing-the-runtime-environment/#exit-codes-and-statuses
const (
	exitSuccess         = 0
	exitInvalidCmdArgs  = 1
	exitK8sConfigErr    = 2
	exitK8sOperationErr = 3
	exitNeutral         = 78
)

var (
	namespace string
)

func newClient() (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, &clientcmd.ConfigOverrides{})

	config, err := loader.ClientConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func createRule(host string, svc string, port int) v1beta1.IngressRule {
	return v1beta1.IngressRule{
		Host: host,
		IngressRuleValue: v1beta1.IngressRuleValue{
			HTTP: &v1beta1.HTTPIngressRuleValue{
				Paths: []v1beta1.HTTPIngressPath{
					{
						Path: "/*",
						Backend: v1beta1.IngressBackend{
							ServiceName: svc,
							ServicePort: intstr.FromInt(port),
						},
					},
				},
			},
		},
	}
}

func addRule(ingress *v1beta1.Ingress, rule v1beta1.IngressRule) (added bool) {
	for _, r := range ingress.Spec.Rules {
		if r.Host == rule.Host {
			return false
		}
	}
	ingress.Spec.Rules = append(ingress.Spec.Rules, rule)
	return true
}

func removeRule(ingress *v1beta1.Ingress, host string) (found bool) {
	for i, r := range ingress.Spec.Rules {
		if r.Host == host {
			ingress.Spec.Rules = append(ingress.Spec.Rules[:i], ingress.Spec.Rules[i+1:]...)
			return true
		}
	}
	return false
}

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
	client, err := newClient()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		return exitK8sConfigErr
	}
	ok := validateCmdArgs(args)
	if !ok {
		return exitInvalidCmdArgs
	}

	operation := args[0]
	ingressName := args[1]
	ingresses := client.ExtensionsV1beta1().Ingresses(namespace)
	ing, err := ingresses.Get(ingressName, metav1.GetOptions{})
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
		rule := createRule(host, serviceName, servicePort)
		added := addRule(ing, rule)
		if !added {
			return exitNeutral
		}
	case "remove":
		host := args[2]
		found := removeRule(ing, host)
		if !found {
			return exitNeutral
		}
	}

	_, err = client.ExtensionsV1beta1().Ingresses(namespace).Update(ing)
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
