/**
 * File              : structs.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 21.12.2020
 * Last Modified Date: 22.01.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */

package server

import (
	PodManager "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/manager"
)

type ServerConfig struct {
	SLACK_API_TOKEN              string
	DOCKER_IMAGE                 string
	MIGRATION_COMMAND            string
	SEED_COMMAND                 string
	SEQUELIZE_MIGRATION_ENV_NAME string
	SEQUELIZE_SEED_ENV_NAME      string
}

type Server struct {
	port    int
	manager PodManager.PodManager
	config  ServerConfig
}

type JobCreationPayload struct {
	Namespace       string            `json:"namespace"`
	JobName         string            `json:"jobName"`
	ConfigMapsNames []string          `json:"configMapsNames"`
	EnvVariablesMap map[string]string `json:"envVariables"`
	DockerImage     string            `json:"dockerImage"`
}
