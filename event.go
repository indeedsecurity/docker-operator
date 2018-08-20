package main

import (
	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

type event struct {
	ServiceName         string   `json:"service_name,omitempty"`
	ServiceDescription  string   `json:"service_description,omitempty"`
	ServiceDeployStatus string   `json:"service_deploy_status,omitempty"`
	Repository          string   `json:"repository,omitempty"`
	LogsURL             string   `json:"logs_url,omitempty"`
	ImageName           string   `json:"image_name,omitempty"`
	Experts             []string `json:"experts,omitempty"`
	Logs                string   `json:"logs,omitempty"`
	ExitCode            string   `json:"exit_code,omitempty"`
	EventType           string   `json:"event_type,omitempty"`
}

func (d *daemon) collectEvents() {
	events, errChan := d.client.Events(d.ctx, types.EventsOptions{})
	for {
		select {
		case e := <-events:
			if e.Type == "container" {
				d.eventContainer(e)
			} else if e.Type == "service" {
				d.eventService(e)
			}
		case err := <-errChan:
			log.Error(err)
		}
	}
}
