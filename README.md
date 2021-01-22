# go-feather-slack-app
## Basic Overview

go-feather-slack app has been created since our DevOps didn't found any relevent application wich was able to do some basic Feather.js/Sequelize migrations, seeds.

With go-feather-slack-app, you'll be able to use some slack commands like ```/migrations``` or ```/seeds``` inside your Slack workspace after installing your App

## Installation

Create a [Slack App](https://api.slack.com/apps) with slach commands and copy your credentials so you will need it as environment variables.

### Environment variables

```bash
SLACK_API_TOKEN=#Your slack application "Signing Secret"
APP_DOCKER_IMAGE=#Your docker image which run the FeatherJs/Sequelize migration
APP_PORT=3030 #Application containerPort (might be used for K8s deployment)
```

## Last Stable Release

## Latest Development Changes


## Usage

## Other Projects
### Inarix migrations

### Inarix seeds

### Inarix T.A.R.S

## Notes
## Contributing

See our [CONTRIBUTING.md](CONTRIBUTING.md)

## Pending Features
