package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	networkingv1 "k8s.io/api/networking/v1"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	clientset, err := getClientset(*kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	globalCSV = InitializeCSV()

	for _, n := range listNamespaces(clientset).Items {
		ns := &Namespace{
			Name:   n.ObjectMeta.Name,
			Labels: n.ObjectMeta.Labels,
		}

		for _, p := range listPods(clientset, ns.Name).Items {
			ns.Pods = append(ns.Pods, &Pod{
				Egress:    []Rule{},
				Labels:    p.ObjectMeta.Labels,
				Ingress:   []Rule{},
				Name:      p.ObjectMeta.Name,
				Namespace: *ns,
			})
		}

		globalCSV.Namespaces = append(globalCSV.Namespaces, ns)
	}

	for _, ns := range globalCSV.Namespaces {
		for _, nwp := range listNetworkPolicies(clientset, ns.Name).Items {
			ns.attachPolicy(nwp.Spec)
		}
	}

	globalCSV.Output()
}

func getClientset(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func listNamespaces(clientset *kubernetes.Clientset) *corev1.NamespaceList {
	ns, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	return ns
}

func listPods(clientset *kubernetes.Clientset, ns string) *corev1.PodList {
	pods, err := clientset.CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	return pods
}

func listNetworkPolicies(clientset *kubernetes.Clientset, ns string) *networkingv1.NetworkPolicyList {
	nwp, err := clientset.NetworkingV1().NetworkPolicies(ns).List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	return nwp
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
