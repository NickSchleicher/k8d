package main

import (
	"flag"
	"os"
	"path/filepath"
)

type options struct {
	kubeConfig *string
	outputFile *string
}

const defaultOutputFile = "drawio.txt"

func getOptions() *options {
	return &options{
		kubeConfig: getKubeConfig(),
		outputFile: flag.String("outputfile", defaultOutputFile, "(optional) draw.io output file name"),
	}
}

func getKubeConfig() *string {
	var kc *string
	if home := homeDir(); home != "" {
		kc = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kc = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	return kc
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
