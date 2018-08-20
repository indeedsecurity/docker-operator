# docker-operator

## features

- sets oom_score on containers
- performs [system prune](https://docs.docker.com/engine/reference/commandline/system_prune/#usage) operations via Docker API
- processes events from the Docker [events API](https://docs.docker.com/engine/api/v1.31/#operation/SystemEvents) and notify owners on Slack based on the `experts` container label (comma separated)

## container labels

These labels are applied to workload containers read by swarm operator to direct and enrich messages when reporting events

```yaml
labels:
  # try to have 2+ experts for each service.
  # these are used for event notifications.
  # specify the email username.
  - "experts=dustind,jdoe"
  - "description="acme service farms chickens"
```

These env vars are applied to swarm-operator so it can connect to the Slack API (optional) and provide enrichement (optional)

```yaml
environment:
    - SLACK_BOT_API_KEY_PATH=/run/secrets/swarm-operator-slack-bot-token
    - SLACK_WORKSPACE_API_KEY_PATH=/run/secrets/swarm-operator-slack-workspace-token
    - SLACK_EMAIL_DOMAIN=indeed.com
    - LOG_URL_FORMAT=https://kibana.internal.net/app/kibana#/discover?query=%[1]s
    - DEPLOY_URL_FORMAT=https://swarm.internal.net/service/ps/%s
```

## restart policy

Set a sensible restart policy so you don't get spammed by alerts if your application is unhealthy

```yaml
restart_policy:
    condition: on-failure
    delay: 5s
    max_attempts: 3
    window: 120s # monitors the process for this period
```