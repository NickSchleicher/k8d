package k8s

import (
	networkingv1 "k8s.io/api/networking/v1"
)

// Namespace contains namespace information and pods within it
type Namespace struct {
	*Cluster `json:"-"`
	Labels   map[string]string
	Name     string
	Pods     []*Pod
}

func (n *Namespace) attachPolicy(spec networkingv1.NetworkPolicySpec) {
	selectors := spec.PodSelector.MatchLabels

	for _, p := range n.Pods {
		// the network policy belongs to this pod
		if len(selectors) == 0 || hasLabels(p.Labels, selectors) {
			p.attachGlobalPolicy(spec)

			// 	attach ingress
			p.Ingress = append(p.Ingress, p.getIngressRules(spec)...)

			//	attach egress
		}
	}
}

func hasLabels(owner map[string]string, required map[string]string) bool {
	for k, r := range required {
		if value, exists := owner[k]; exists {
			if value != r {
				return false
			}
		} else {
			return false
		}
	}

	return true
}
