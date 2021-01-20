/**
 * File              : structs.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 21.12.2020
 * Last Modified Date: 20.01.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */

package server

import (
	PodManager "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/manager"
)

type Server struct {
	port            int
	manager         PodManager.PodManager
	SLACK_API_TOKEN string
	DOCKER_IMAGE    string
}

type JobCreationPayload struct {
	Namespace       string   `json:"namespace"`
	JobName         string   `json:"jobName"`
	ConfigMapsNames []string `json:"configMapsNames"`
	DockerImage     string   `json:"dockerImage"`
	CleanUp         bool     `json:"cleanup"`
}

type PodsRequestPayload struct {
	Namespace string `json:"namespace"`
}

type PodRequestPayload struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}
