/**
 * File              : structs.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 21.12.2020
 * Last Modified Date: 04.02.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */

package server

import (
	PodManager "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/manager"
	"github.com/slack-go/slack"
)

//ServerConfig is the basic server required configuration
type ServerConfig struct {
	SLACK_API_TOKEN              string
	SLACK_SIGNING_SECRET         string
	SLACK_ANSWER_CHANNEL_ID      string
	DOCKER_IMAGE                 string
	MIGRATION_COMMAND            string
	SEED_COMMAND                 string
	SEQUELIZE_MIGRATION_ENV_NAME string
	SEQUELIZE_SEED_ENV_NAME      string
}

type Server struct {
	port        int
	manager     PodManager.PodManager
	config      ServerConfig
	slackClient slack.Client
}

type JobCreationPayload struct {
	Namespace       string            `json:"namespace"`
	JobName         string            `json:"jobName"`
	ConfigMapsNames []string          `json:"configMapsNames"`
	EnvVariablesMap map[string]string `json:"envVariables"`
	DockerImage     string            `json:"dockerImage"`
}

type SlackApiEventPayload struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}
