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
APP_DOCKER_IMAGE: 894517829775.dkr.ecr.eu-west-1.amazonaws.com/inarix-api
APP_MIGRATION_COMMAND: #Used command to trigger migration creation
APP_SEED_COMMAND: #Used command to trigger seed creation
APP_SEQUELIZE_MIGRATION_ENV_NAME: #Name of the environment key which will have the selected migration name
APP_SEQUELIZE_SEED_ENV_NAME: #Name of the environment key which will have the selected migration name
GOENV: production
```

## Last Stable Release

See [SECURITY.md](SECURITY.md)

## Usage

Use this go application to be able to launch migration and seeds with a simple slack slach command !

![Migration Creation GIF]()

![Seed Creation GIF]()
### Inarix T.A.R.S

![T.A.R.S Simple response]()

## Notes
## Contributing

See our [CONTRIBUTING.md](CONTRIBUTING.md)
