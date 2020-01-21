package main

import (
	"fmt"
	"log"

	"github.com/NickSchleicher/k8d/internal/draw"
	"github.com/NickSchleicher/k8d/internal/k8s"
)

const uploadInstructions = `*********************
Thanks for using k8d, your file has been saved to %s
Go to www.draw.io
Create New Diagram
Blank Diagram > Create
Arrange > Insert > Advanced > CSV
Copy the contents from %s and paste it into the text box
Import
`

func main() {
	opts := getOptions()

	cluster, err := k8s.Init(opts.kubeConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = cluster.BuildNetworkConnections()
	if err != nil {
		log.Fatal(err)
	}

	err = draw.Output(cluster, opts.outputFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(uploadInstructions, *opts.outputFile, *opts.outputFile)
}
