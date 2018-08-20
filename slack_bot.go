package main

import (
	"strings"

	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

func (d *daemon) sendSlackMessage(experts []string, logs string, options ...slack.MsgOption) {
	var users []string
	if len(experts) == 1 && experts[0] == "" {
		log.Info("No experts: %v", experts)
		return
	}
	for _, user := range experts {
		if user != "" {
			if d.slackUsers[user] == "" {
				// update cache if not found
				if !strings.Contains(user, "@") {
					if d.config.slackEmailDomain == "" {
						log.Error("Full email not provided for expert and SLACK_EMAIL_DOMAIN is empty. Cannot look up user in slack.")
					}
					user = user + "@" + d.config.slackEmailDomain
				}
				userInfo, err := d.slackWorkspace.GetUserByEmail(user)
				if err != nil {
					log.Errorf("Could not find user %s\n", user)
					continue
				}
				d.slackUsers[user] = userInfo.ID
			}
			users = append(users, d.slackUsers[user])
		}
	}
	var chID string
	if len(users) > 1 {
		// open a conversation with multiple users
		channel := slack.OpenConversationParameters{
			Users: users,
		}
		openChannel, _, _, err := d.slackBot.OpenConversation(&channel)
		if err != nil {
			log.Errorf("Error opening channel: %s\n", err)
			return
		}
		chID = openChannel.ID
	} else if len(users) == 1 {
		// Instant message for single users
		_, _, openChannel, chErr := d.slackBot.OpenIMChannel(users[0])
		if chErr != nil {
			log.Errorf("Error opening IM: %s\n", chErr)
			return
		}
		chID = openChannel
	} else {
		// No users, don't send a message
		if len(experts) > 0 {
			log.Errorf("Error getting slack IDs for users: %s", experts)
		}
		return
	}

	_, _, _, err := d.slackBot.SendMessage(chID, options...)
	if err != nil {
		log.Errorf("Error Sending Message: %s\n", err)
	}
	if len(logs) > 0 {
		params := slack.FileUploadParameters{
			Title:    "log output",
			Filetype: "txt",
			Content:  logs,
			Channels: []string{chID},
		}
		_, err := d.slackBot.UploadFile(params)
		if err != nil {
			log.Errorf("Error uploading logs: %s\n", err)
		}
	}
}
