package main

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/nlopes/slack"

	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

const defaultOOMScore = 500
const defaultOOMerTick = 60
const defaultPruneTick = 1800

type config struct {
	logURLFormat     string
	deployURLFormat  string
	slackEmailDomain string
}

type daemon struct {
	config               config
	client               *client.Client
	signals              chan os.Signal
	signalOOMer          chan os.Signal
	signalPruner         chan os.Signal
	signalEventCollector chan os.Signal
	ctx                  context.Context
	cancelCtx            context.CancelFunc
	slackEnabled         bool
	slackBot             *slack.Client
	slackWorkspace       *slack.Client
	slackUsers           map[string]string
}

func newDaemon() *daemon {
	log.SetFormatter(&log.JSONFormatter{})
	slackBotAPIKeyPath := os.Getenv("SLACK_BOT_API_KEY_PATH")
	slackWorkspaceAPIKeyPath := os.Getenv("SLACK_WORKSPACE_API_KEY_PATH")
	slackEmailDomain := os.Getenv("SLACK_EMAIL_DOMAIN")
	logURLFormat := os.Getenv("LOG_URL_FORMAT")
	deployURLFormat := os.Getenv("DEPLOY_URL_FORMAT")

	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	d := daemon{
		config: config{
			logURLFormat:    logURLFormat,
			deployURLFormat: deployURLFormat,
		},
		client:       cli,
		ctx:          ctx,
		cancelCtx:    cancel,
		signals:      make(chan os.Signal),
		slackEnabled: false,
		slackUsers:   make(map[string]string),
		// signalOOMer:          make(chan os.Signal),
		// signalEventCollector: make(chan os.Signal),
	}

	if len(slackBotAPIKeyPath) > 0 && len(slackWorkspaceAPIKeyPath) > 0 {
		botAPIKey, err := ioutil.ReadFile(slackBotAPIKeyPath)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		botAPI := slack.New(strings.TrimSpace(string(botAPIKey)))

		workspaceAPIKey, err := ioutil.ReadFile(slackWorkspaceAPIKeyPath)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		workspaceAPI := slack.New(strings.TrimSpace(string(workspaceAPIKey)))

		if botAPI != nil && workspaceAPI != nil {
			d.slackBot = botAPI
			d.slackWorkspace = workspaceAPI
			d.config.slackEmailDomain = slackEmailDomain
			d.slackEnabled = true
		}
	}

	return &d
}

func (d *daemon) quit() {
	<-d.signals
	log.Info("shutting down")
	// d.signalOOMer <- s
	// d.signalPruner <- s
	// d.signalEventCollector <- s
	d.cancelCtx()
}

func main() {

	d := newDaemon()
	defer d.quit()

	signal.Notify(d.signals, os.Interrupt)

	oomTicker := time.NewTicker(defaultOOMerTick * time.Second)
	go func() {
		d.applyOOMScoreOnContainers()
		for {
			select {
			case <-oomTicker.C:
				d.applyOOMScoreOnContainers()
			case <-d.signalOOMer:
				oomTicker.Stop()
				return
			}
		}
	}()

	pruneTicker := time.NewTicker(defaultPruneTick * time.Second)
	go func() {
		d.pruneDockerSystem()
		for {
			select {
			case <-pruneTicker.C:
				d.pruneDockerSystem()
			case <-d.signalPruner:
				pruneTicker.Stop()
				return
			}
		}
	}()
	go d.collectEvents()
}
