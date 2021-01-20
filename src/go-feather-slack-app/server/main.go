/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 09.12.2020
 * Last Modified Date: 20.01.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	PodManager "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/manager"
	"github.com/slack-go/slack"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (self *Server) fromSlackTextToStruct(slackTextArguments []string, structHandler *JobCreationPayload) error {
	if len(slackTextArguments) == 0 {
		return errors.New("No slack argument found !")
	} else if len(slackTextArguments) < 2 {
		return errors.New("Not enough slack argument found !")
	}
	structHandler.DockerImage = self.DOCKER_IMAGE + slackTextArguments[0]
	structHandler.Namespace = "default"
	structHandler.JobName = "go-feather-slack-app-"

	for index, slackTextArguments := range slackTextArguments[1:] {
		log.Printf("Argument %d -> %s", index, slackTextArguments)
	}
	return nil
}

func (self *Server) Pods() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendStatusMethodNotAllowed(w)
			return
		}
		var FormValues PodsRequestPayload
		if err := fromBodyToStruct(r.Body, &FormValues); err != nil {
			log.Printf("An error occured while unmarchalling your payload : %s", err.Error())
			sendStatusInternalError(w)
			return
		}

		log.Printf("Selected namespace for pod introspection is %s", FormValues.Namespace)
		pods, err := self.manager.GetPods(FormValues.Namespace)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err.Error())
		}
		marchalledKubernetesItems, err := json.Marshal(pods.Items)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err.Error())
		}
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintln(w, string(marchalledKubernetesItems))
	}
}

func (self *Server) CreateJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendStatusMethodNotAllowed(w)
			return
		}
	}
}

func (self *Server) GetPod() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendStatusMethodNotAllowed(w)
			return
		}

		var FormValues JobCreationPayload
		if err := fromBodyToStruct(r.Body, &FormValues); err != nil {
			log.Printf("An error occured while unmarchalling your payload : %s", err.Error())
			sendStatusInternalError(w)
			return
		}

	}
}

func (self *Server) SubmitJobCreation(slackTextArguments []string, w http.ResponseWriter, r *http.Request) string {
	var FormValues JobCreationPayload
	if err := self.fromSlackTextToStruct(slackTextArguments, &FormValues); err != nil {
		log.Printf("An error occured while unmarchalling your payload : %s", err.Error())
		sendStatusInternalError(w)
		return ""
	}
	configMapRefs := self.manager.CreateConfigRefSpec(FormValues.ConfigMapsNames)
	prefixName := FormValues.JobName + "-job"
	jobSpec := self.manager.CreateJobSpec("go-feather-slack-app-job", prefixName, FormValues.DockerImage, nil, configMapRefs)

	pod, err := self.manager.CreateJob(FormValues.Namespace, prefixName, *jobSpec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error during creation of Job : %s", err.Error())
	}
	logs, err := self.manager.GetPodLogs(FormValues.Namespace, pod)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error during creation of Job : %s", err.Error())
	}

	if FormValues.CleanUp {
		log.Printf("Cleaning up %s : ", pod.Labels["job-name"])
		self.manager.DeleteJob(FormValues.Namespace, pod.Labels["job-name"])
		self.manager.DeletePod(FormValues.Namespace, pod.GetName())
	}
	return logs

}

func (self *Server) handleSlackCommand() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendStatusMethodNotAllowed(w)
			return
		}
		verifier, err := slack.NewSecretsVerifier(r.Header, self.SLACK_API_TOKEN)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = verifier.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch s.Command {
		case "/echo":
			SendSlackMessage(s.UserName+" typed -> "+s.Text+" on channel "+s.ChannelName, w)
		case "/migration":
			slackTextArguments := strings.Fields(s.Text)
			err := r.ParseForm()
			if err != nil {
				log.Println("ERROR IN PARSING FORM")
			}

			log.Println("action param => ", slackTextArguments)
			version := slackTextArguments[0]
			versionRegex, _ := regexp.Compile("v[0-9]+\\.[0-9]+\\.[0-9]+")
			hasVersionSpecified := versionRegex.MatchString(version)

			if !hasVersionSpecified {
				SendSlackMessage("You must specify a good version (eg. v.1.0.0) : "+version, w)
			}
			self.SubmitJobCreation(slackTextArguments, w, r)

			//var payload string
			//messageToSend := &slack.Msg{Text: payload}
			//b, err := json.Marshal(messageToSend)
			//if err != nil {
			//	w.WriteHeader(http.StatusInternalServerError)
			//	return
			//}
			//w.Header().Set("Content-Type", "application/json")
			//w.Write(b)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func New(listenPort int, podManager PodManager.PodManager) *Server {
	SLACK_API_TOKEN := os.Getenv("SLACK_API_TOKEN")
	DOCKER_IMAGE := os.Getenv("APP_DOCKER_IMAGE")
	if SLACK_API_TOKEN == "" || DOCKER_IMAGE == "" {
		log.Panicln(errors.New("Environment variables DOCKER_IMAGE or SLACK_API_TOKEN are missing").Error())
	}
	return &Server{port: listenPort, manager: podManager, SLACK_API_TOKEN: SLACK_API_TOKEN, DOCKER_IMAGE: DOCKER_IMAGE}
}

func Listen(manager PodManager.PodManager) {

	appPortStr := os.Getenv("APP_PORT")
	if appPortStr == "" {
		log.Panicln(errors.New("Environment variable APP_PORT is missing").Error())
	}
	appPort, err := strconv.Atoi(appPortStr)
	if err != nil {
		log.Panicln(err.Error())
	}
	server := New(appPort, manager)
	http.HandleFunc("/", server.handleSlackCommand())
	http.HandleFunc("/pods", server.Pods())
	http.HandleFunc("/pod", server.GetPod())
	http.HandleFunc("/healthz", healthz)

	log.Println("Server started on port " + appPortStr)
	if err := http.ListenAndServe(":"+appPortStr, nil); err != nil {
		panic(err.Error())
	}
}
