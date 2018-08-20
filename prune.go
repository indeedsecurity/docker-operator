package main

import (
	"github.com/docker/docker/api/types/filters"
	log "github.com/sirupsen/logrus"
)

func (d *daemon) pruneDockerSystem() {
	_, err := d.client.ContainersPrune(d.ctx, filters.Args{})
	if err != nil {
		log.Error(err)
	}

	_, err = d.client.NetworksPrune(d.ctx, filters.Args{})
	if err != nil {
		log.Error(err)
	}

	_, err = d.client.ImagesPrune(d.ctx, filters.Args{})
	if err != nil {
		log.Error(err)
	}

	_, err = d.client.VolumesPrune(d.ctx, filters.Args{})
	if err != nil {
		log.Error(err)
	}
}
