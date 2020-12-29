/**
 * File              : structs.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 21.12.2020
 * Last Modified Date: 29.12.2020
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */

package server

import (
	PodManager "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/manager"
)

//const SLACK_API_TOKEN string := os.Getenv("SLACK_API_TOKEN")
//TODO : add SLACK_API_TOKEN !

type Server struct {
	port            int
	manager         PodManager.PodManager
	SLACK_API_TOKEN string
}

type JobCreationPayload struct {
	Namespace       string   `json:"namespace"`
	JobName         string   `json:"jobName"`
	ConfigMapsNames []string `json:"configMapsNames"`
	DockerImage     string   `json:"dockerImage"`
}

type PodsRequestPayload struct {
	Namespace string `json:"namespace"`
}
