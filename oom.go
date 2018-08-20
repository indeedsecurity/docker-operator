package main

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

func (d *daemon) applyOOMScoreOnContainers() {
	containers, err := d.client.ContainerList(d.ctx, types.ContainerListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, container := range containers {
		inspect, err := d.client.ContainerInspect(d.ctx, container.ID)
		if err != nil {
			log.Fatal(err)
		}

		pid := inspect.State.Pid
		var pidStr string
		if pid == 0 {
			pidStr = "self"
		} else {
			pidStr = strconv.Itoa(pid)
		}

		maxTries := 2
		oomScoreAdj := defaultOOMScore
		oomScoreAdjPath := path.Join("/host/proc", pidStr, "oom_score_adj")
		if !pathExists(oomScoreAdjPath) {
			oomScoreAdjPath = path.Join("/proc", pidStr, "oom_score_adj")
		}
		for i := 0; i < maxTries; i++ {
			err = ioutil.WriteFile(oomScoreAdjPath, []byte(strconv.Itoa(oomScoreAdj)), 0700)
			if err != nil {
				if os.IsNotExist(err) {
					log.Infof("%s does not exist", oomScoreAdjPath)
				}

				log.Info(err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		if err != nil {
			log.Infof("failed to set %s to %d on container %s: %v", oomScoreAdjPath, oomScoreAdj, inspect.Name, err)
		}
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
