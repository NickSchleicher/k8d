package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Cluster defines how to interact with a kubernetes cluster
type Cluster struct {
	client     *kubernetes.Clientset
	configPath *string
	Namespaces []*Namespace
}

// BuildNetworkConnections maps out what object has restrictions on where they can connect
func (c *Cluster) BuildNetworkConnections() error {
	err := c.addNamespaces()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cluster) addNamespaces() error {
	namespaceList, err := c.client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, n := range namespaceList.Items {
		ns := &Namespace{
			Cluster: c,
			Labels:  n.ObjectMeta.Labels,
			Name:    n.ObjectMeta.Name,
		}

		pods, err := c.getPods(ns)
		if err != nil {
			return err
		}
		ns.Pods = pods

		c.Namespaces = append(c.Namespaces, ns)
	}

	for _, ns := range c.Namespaces {
		err := c.attachNetworkPolicies(ns)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) getPods(ns *Namespace) ([]*Pod, error) {
	podList, err := c.client.CoreV1().Pods(ns.Name).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var pods []*Pod
	for _, p := range podList.Items {
		pods = append(pods, &Pod{
			Cluster:   c,
			Labels:    p.ObjectMeta.Labels,
			Name:      p.ObjectMeta.Name,
			Namespace: ns,
		})
	}

	return pods, nil
}

func (c *Cluster) attachNetworkPolicies(ns *Namespace) error {
	policyList, err := c.client.NetworkingV1().NetworkPolicies(ns.Name).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, p := range policyList.Items {
		ns.attachPolicy(p.Spec)
	}

	return nil
}

// Init creates a way to connect with a cluster
func Init(path *string) (*Cluster, error) {
	c := &Cluster{
		configPath: path,
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", *path)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	c.client = client

	return c, nil
}
