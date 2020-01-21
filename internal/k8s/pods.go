package k8s

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Pod contains pod information and the objects it can ingress or egress to
type Pod struct {
	*Cluster `json:"-"`
	Egress   []Rule
	GlobalPolicy
	Ingress   []Rule
	Labels    map[string]string
	Name      string
	Namespace *Namespace `json:"-"`
}

func (p *Pod) getIngressRules(spec networkingv1.NetworkPolicySpec) []Rule {
	var rules []Rule
	for _, i := range spec.Ingress {
		var ports []Port

		for _, port := range i.Ports {
			ports = append(ports, Port{
				Value:    port.Port.StrVal,
				Protocol: string(*port.Protocol),
			})
		}

		if len(i.From) == 0 {
			rules = append(rules, Rule{
				Everywhere: true,
				Ports:      ports,
			})
		} else {
			for _, f := range i.From {
				rule := Rule{
					IPBlock: f.IPBlock,
				}

				rule.Pods = p.getAffectedPods(f.NamespaceSelector, f.PodSelector)

				rules = append(rules, rule)
			}
		}
	}

	return rules
}

// Rule contains the ingress or egress configuration
type Rule struct {
	Everywhere bool
	IPBlock    *networkingv1.IPBlock
	Pods       []*Pod
	Ports      []Port
}

// Port contains the port number and protocol
type Port struct {
	Protocol string
	Value    string
}

// GlobalPolicy contains the possible default policies
type GlobalPolicy struct {
	AllowEgress  bool
	AllowIngress bool
	DenyEgress   bool
	DenyIngress  bool
}

func (p *Pod) attachGlobalPolicy(spec networkingv1.NetworkPolicySpec) {
	for _, pt := range spec.PolicyTypes {
		if pt == "Ingress" {
			count := len(spec.Ingress)
			if count == 0 {
				p.GlobalPolicy.DenyIngress = true
			} else if count == 1 {
				if len(spec.Ingress[0].From) == 0 && len(spec.Ingress[0].Ports) == 0 {
					p.GlobalPolicy.AllowIngress = true
				}
			}
		} else if pt == "Egress" {
			count := len(spec.Egress)
			if count == 0 {
				p.GlobalPolicy.DenyEgress = true
			} else if count == 1 {
				if len(spec.Egress[0].To) == 0 && len(spec.Egress[0].Ports) == 0 {
					p.GlobalPolicy.AllowEgress = true
				}
			}
		}
	}
}

func (p *Pod) getAffectedPods(nsSelector, podSelector *metav1.LabelSelector) []*Pod {
	var nsLabels map[string]string
	var podLabels map[string]string

	if nsSelector != nil {
		nsLabels = nsSelector.MatchLabels
	}
	if podSelector != nil {
		podLabels = podSelector.MatchLabels
	}

	var pods []*Pod
	for _, ns := range p.Cluster.Namespaces {
		if nsSelector == nil && p.Namespace == ns || nsSelector != nil && hasLabels(ns.Labels, nsLabels) {
			for _, pod := range ns.Pods {
				if p != pod && hasLabels(p.Labels, podLabels) {
					fmt.Println(p.Name, podLabels)
					pods = append(pods, pod)
				}
			}
		}
	}

	return pods
}
