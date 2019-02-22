package main

import (
	"flag"
	"log"
	"os"

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
	namespace   string
	ingressHost string
	ingressName string
	serviceName string
	servicePort int
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

func addRule(ingress *v1beta1.Ingress) {
	rule := v1beta1.IngressRule{
		Host: ingressHost,
		IngressRuleValue: v1beta1.IngressRuleValue{
			HTTP: &v1beta1.HTTPIngressRuleValue{
				Paths: []v1beta1.HTTPIngressPath{
					{
						Path: "/*",
						Backend: v1beta1.IngressBackend{
							ServiceName: serviceName,
							ServicePort: intstr.FromInt(servicePort),
						},
					},
				},
			},
		},
	}
	ingress.Spec.Rules = append(ingress.Spec.Rules, rule)
}

func removeRule(ingress *v1beta1.Ingress) {
	for i, r := range ingress.Spec.Rules {
		if r.Host == ingressHost {
			ingress.Spec.Rules = append(ingress.Spec.Rules[:i], ingress.Spec.Rules[i+1:]...)
			break
		}
	}
}

func run(operation string) int {
	client, err := newClient()
	if err != nil {
		return exitK8sConfigErr
	}

	ingresses := client.ExtensionsV1beta1().Ingresses(namespace)
	ing, err := ingresses.Get(ingressName, metav1.GetOptions{})
	if err != nil {
		return exitK8sOperationErr
	}

	switch operation {
	case "add":
		addRule(ing)
	case "remove":
		removeRule(ing)
	default:
		return exitInvalidCmdArgs
	}

	updated, err := client.ExtensionsV1beta1().Ingresses(namespace).Update(ing)
	if err != nil {
		return exitK8sOperationErr
	}

	log.Printf("updated: %#v\n", updated)
	return exitSuccess
}

func main() {
	flag.StringVar(&namespace, "namespace", "default", "kubernetes namespace")
	flag.StringVar(&ingressHost, "host", "", "kubernetes ingress host")
	flag.StringVar(&ingressName, "name", "", "kubernetes ingress name")
	flag.StringVar(&serviceName, "svcname", "", "kubernetes service name")
	flag.IntVar(&servicePort, "svcport", 80, "port number of kubernetes service")
	flag.Parse()

	op := flag.Arg(0)
	status := run(op)
	os.Exit(status)
}
