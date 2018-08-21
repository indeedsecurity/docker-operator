package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

func (d *daemon) eventContainer(e events.Message) {
	if e.Status == "die" || e.Status == "oom" || e.Status == "health_status" {

		exitCode := e.Actor.Attributes["exitCode"]
		// 137 exit code is if Docker needs to sigkill following entrypoint
		// not responding to sigterm (128 + 9 = 137)
		// 143 is sigterm (128 + 15 = 143)
		if exitCode == "0" || exitCode == "137" || exitCode == "143" {
			return
		}

		// TODO - health status events
		if e.Status == "health_status" {
			spew.Dump(e)
		}

		ev := event{
			ServiceName:        e.Actor.Attributes["com.docker.swarm.service.name"],
			ServiceDescription: e.Actor.Attributes["description"],
			Repository:         e.Actor.Attributes["repository"],
			ImageName:          e.Actor.Attributes["image"],
			Experts:            strings.Split(e.Actor.Attributes["experts"], ","),
			ExitCode:           e.Actor.Attributes["exitCode"],
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

		logReader, err := d.client.ContainerLogs(d.ctx, e.Actor.ID, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
		})
		if err != nil {
			log.Error(err)
		} else {
			logs, err := ioutil.ReadAll(io.LimitReader(logReader, 1024*1024*100)) // 100M memory
			if err == nil {
				logLines := strings.Split(string(logs), "\n")
				logLen := len(logLines)
				numLines := 300
				if logLen > numLines {
					ev.Logs = strings.Join(logLines[logLen-numLines:], "\n")
				} else {
					ev.Logs = strings.Join(logLines, "\n")
				}
			} else {
				log.Error(err)
			}
		}

		var text string
		if len(ev.ServiceName) > 1 {
			text += "*Service Name:* " + ev.ServiceName + "\n"
			text += "*Service Description:* " + ev.ServiceDescription + "\n"
			text += "*Service Deploy Status:* " + ev.ServiceDeployStatus + "\n"
		}
		if len(ev.Repository) > 1 {
			text += "*Repository:* " + ev.Repository + "\n"
		}
		text += "*Image Name:* " + ev.ImageName + "\n"
		text += "*Event Type:* " + ev.EventType + "\n"
		text += "*Exit Code:* " + ev.ExitCode + "\n"
		var msgOptions []slack.MsgOption
		msgOptions = append(msgOptions, slack.MsgOptionText(text, false))

		// this dumb option makes it use the name you give it rather than just "bot"
		msgOptions = append(msgOptions, slack.MsgOptionAsUser(true))

		evJSON, err := json.Marshal(ev)
		if err != nil {
			log.Error(err)
		}
		fmt.Printf("%s\n", evJSON)
		if d.slackEnabled {
			d.sendSlackMessage(ev.Experts, ev.Logs, msgOptions...)
		}
	}
}
