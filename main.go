package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), `Usage:
  %s add -ingress=<INGRESS_NAME> -host=<INGRESS_HOST> -service=<SERVICE_NAME> -port=<SERVICE_PORT>
  %s remove -ingress=<INGRESS_NAME> -host=<INGRESS_HOST>

Options:
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

func run(operation string) int {
	client, err := initClient()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		return exitK8sOperationErr
	}
	ingress, err := client.ExtensionsV1beta1().Ingresses(namespace).Get(ingressName, metav1.GetOptions{})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		return exitK8sOperationErr
	}

	switch operation {
	case "add":
		rule := createRule()
		if added := addRule(ingress, rule); !added {
			return exitSuccess
		}
	case "remove":
		if found := removeRule(ingress); !found {
			return exitSuccess
		}
	}
	_, err = client.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
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

func initClient() (client *kubernetes.Clientset, err error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	)
	config, err := loader.ClientConfig()
	if err != nil {
		return nil, err
	}
	if namespace == "" {
		// Use namespace in context if it's empty.
		namespace, _, err = loader.Namespace()
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(config)
}

func createRule() v1beta1.IngressRule {
	return v1beta1.IngressRule{
		Host: ingressHost,
		IngressRuleValue: v1beta1.IngressRuleValue{
			HTTP: &v1beta1.HTTPIngressRuleValue{
				Paths: []v1beta1.HTTPIngressPath{
					{
						Path: pathRule,
						Backend: v1beta1.IngressBackend{
							ServiceName: serviceName,
							ServicePort: intstr.FromInt(servicePort),
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

func removeRule(ingress *v1beta1.Ingress) (found bool) {
	for i, r := range ingress.Spec.Rules {
		if r.Host == ingressHost {
			ingress.Spec.Rules = append(ingress.Spec.Rules[:i], ingress.Spec.Rules[i+1:]...)
			return true
		}
	}
	return false
}
