package ingress

import (
	"k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func New(ingress string, opt *Option) (*RuleEditor, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	)

	config, err := loader.ClientConfig()
	if err != nil {
		return nil, err
	}

	namespace, _, err := loader.Namespace()
	if err != nil {
		return nil, err
	}
	if opt != nil && opt.Namespace != "" {
		namespace = opt.Namespace
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	ing, err := client.ExtensionsV1beta1().Ingresses(namespace).Get(ingress, opt.GetOptions)
	if err != nil {
		return nil, err
	}

	return &RuleEditor{
		Namespace: namespace,
		Client:    client,
		Ingress:   ing,
	}, nil
}

type Option struct {
	Namespace  string
	GetOptions v1.GetOptions
}

type RuleEditor struct {
	Client    *kubernetes.Clientset
	Namespace string
	Ingress   *v1beta1.Ingress
}

func (e *RuleEditor) Add(host string, svc string, port int) (added bool, err error) {
	rule := v1beta1.IngressRule{
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
	added = addRule(e.Ingress, rule)
	if !added {
		return false, nil
	}
	err = apply(e.Client, e.Namespace, e.Ingress)
	return true, err
}

func (e *RuleEditor) Remove(host string) (found bool, err error) {
	found = removeRule(e.Ingress, host)
	if !found {
		return false, nil
	}
	err = apply(e.Client, e.Namespace, e.Ingress)
	return true, err
}

func apply(client *kubernetes.Clientset, namespace string, ingress *v1beta1.Ingress) error {
	_, err := client.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
	return err
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
