//go:build ignore
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"example.com/project/workflows"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	defer c.Close()

	w := worker.New(c, "batch-processing", worker.Options{})

	w.RegisterWorkflow(workflows.BatchProcessWorkflow)
	w.RegisterActivity(workflows.ExtractActivity)
	w.RegisterActivity(workflows.TransformActivity)
	w.RegisterActivity(workflows.LoadActivity)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker:", err)
	}
}
