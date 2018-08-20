package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/filters"
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
)

func (d *daemon) eventService(e events.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	opts := types.NodeListOptions{}
	nodes, err := d.client.NodeList(ctx, opts)
	if err != nil {
		fmt.Printf("Error listing nodes: %s", err)
	}
	hostname, _ := os.Hostname()
	var isLeader = false
	// Search through all nodes and determine if current host is manager leader
	for _, node := range nodes {
		if node.ManagerStatus.Leader && node.Description.Hostname == hostname {
			isLeader = true
		}
	}
	if !isLeader {
		cancel()
		return
	}
	// Get container with same name as the container that generated the event
	// The event contains much less info than a container event
	containers, err := d.client.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "name",
			Value: e.Actor.Attributes["name"],
		}),
	})
	if err != nil {
		log.Error(err)
	}
	if len(containers) > 0 {
		// There can be multiple replicas, but this information will be consistent across them all
		container := containers[0]
		if container.Labels["com.docker.swarm.service.name"] == e.Actor.Attributes["name"] {
			ev := event{
				ServiceName:        e.Actor.Attributes["com.docker.swarm.service.name"],
				ServiceDescription: e.Actor.Attributes["description"],
				Repository:         e.Actor.Attributes["repository"],
				ImageName:          container.Image,
				Experts:            strings.Split(container.Labels["experts"], ","),
				ExitCode:           "",
				EventType:          e.Status,
			}

			if len(ev.ServiceName) > 0 {
				if len(d.config.logURLFormat) > 0 {
					ev.LogsURL = fmt.Sprintf(d.config.logURLFormat, ev.ServiceName)
				}
				if len(d.config.deployURLFormat) > 0 {
					ev.ServiceDeployStatus = fmt.Sprintf(d.config.deployURLFormat, ev.ServiceName)
				}
			}

			evJSON, err := json.Marshal(ev)
			if err != nil {
				log.Error(err)
			}
			fmt.Printf("%s\n", evJSON)
		}
	}
	cancel()
}
