//go:build ignore
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"example.com/project/internal/activities"
	"example.com/project/internal/workflows"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	defer c.Close()

	w := worker.New(c, "realtime-workflow", worker.Options{})

	w.RegisterWorkflow(workflows.ProcessEventWorkflow)
	w.RegisterActivity(activities.ValidateEvent)
	w.RegisterActivity(activities.ProcessEvent)
	w.RegisterActivity(activities.NotifyDownstream)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker:", err)
	}
}
