/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 09.12.2020
 * Last Modified Date: 22.01.2021
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
	"time"

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

	dockerTag := slackTextArguments[0]

	structHandler.DockerImage = self.config.DOCKER_IMAGE + ":" + dockerTag
	structHandler.Namespace = "default"
	structHandler.JobName = "go-feather-slack-app-" + strconv.Itoa(int(time.Now().Unix()))
	structHandler.ConfigMapsNames = slackTextArguments[1:]
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
			log.Println("error : ", err.Error())
			SendSlackMessage("An error occured while unmarchalling your payload : "+err.Error(), w)
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

func (self *Server) SubmitJobCreation(commandName string, slackTextArguments []string, w http.ResponseWriter, r *http.Request) {
	var FormValues JobCreationPayload
	if err := self.fromSlackTextToStruct(slackTextArguments, &FormValues); err != nil {
		SendSlackMessage("An error occured while unmarchalling your payload : "+err.Error(), w)
	}
	configMapRefs := self.manager.CreateConfigRefSpec(FormValues.ConfigMapsNames)
	prefixName := FormValues.JobName + "-job"
	jobSpec := self.manager.CreateJobSpec("go-feather-slack-app-job", prefixName, FormValues.DockerImage, nil, configMapRefs)

	log.Println("Creating ", FormValues.JobName)
	pod, err := self.manager.CreateJob(FormValues.Namespace, prefixName, *jobSpec)

	if err != nil {
		log.Printf("Error during creation of Job: %s", err.Error())
		SendSlackMessage("Error during creation of Job: "+err.Error(), w)
		return
	}

	SendSlackMessage("Successfully created "+pod.Name, w)
}

func (self *Server) FetchJobPodLogs(podNamespace string, podName string, w http.ResponseWriter) {
	logs, err := self.manager.GetPodLogs(podNamespace, podName)
	if err != nil {
		log.Printf("Error when fetching Job logs: %s", err.Error())
		SendSlackMessage("Error when fetching Job logs: "+err.Error(), w)
		return
	}
	SendSlackMessage(logs, w)
}

func (self *Server) handleSlackCommand() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendStatusMethodNotAllowed(w)
			return
		}
		verifier, err := slack.NewSecretsVerifier(r.Header, self.config.SLACK_API_TOKEN)
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
		log.Printf("Config is : {MIGRATION_COMMAND: %s, SEED_COMMAND: %s}", self.config.MIGRATION_COMMAND, self.config.SEED_COMMAND)

		switch s.Command {
		case "/echo":
			SendSlackMessage(s.UserName+" typed -> "+s.Text+" on channel "+s.ChannelName, w)
		case self.config.SEED_COMMAND:
			slackTextArguments := strings.Fields(s.Text)
			err := r.ParseForm()
			if err != nil {
				SendSlackMessage("Error : "+err.Error(), w)
				log.Println("[ERROR] " + err.Error())
				return
			}

			if len(slackTextArguments) < 2 {
				SendSlackMessage("You must at least specify a version and a seed name !", w)
				return
			}

			version := slackTextArguments[0]
			versionRegex, _ := regexp.Compile("v[0-9]+\\.[0-9]+\\.[0-9]+")
			hasVersionSpecified := versionRegex.MatchString(version)

			if !hasVersionSpecified {
				SendSlackMessage("You must specify a good version (eg. v.1.0.0) : "+version, w)
				return
			}

			self.SubmitJobCreation(s.Command, slackTextArguments, w, r)
			return
		case self.config.MIGRATION_COMMAND:
			slackTextArguments := strings.Fields(s.Text)
			err := r.ParseForm()
			if err != nil {
				SendSlackMessage("Error : "+err.Error(), w)
				log.Println("[ERROR] " + err.Error())
				return
			}

			if len(slackTextArguments) < 1 {
				SendSlackMessage("You must at least specify a version !", w)
				return
			}

			version := slackTextArguments[0]
			versionRegex, _ := regexp.Compile("v[0-9]+\\.[0-9]+\\.[0-9]+")
			hasVersionSpecified := versionRegex.MatchString(version)

			if !hasVersionSpecified {
				SendSlackMessage("You must specify a good version (eg. v.1.0.0) : "+version, w)
				return
			}
			self.SubmitJobCreation(s.Command, slackTextArguments, w, r)
			return
		default:
			SendSlackMessage("Current slack command is not implemented yet !", w)
			return
		}

	}
}

func New(listenPort int, podManager PodManager.PodManager) *Server {
	appConfig := initConfig()
	return &Server{port: listenPort, manager: podManager, config: *appConfig}
}

func Listen(manager PodManager.PodManager) {

	log.SetFlags(log.Lmsgprefix | log.LstdFlags)
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
	http.HandleFunc("/healthz", healthz)

	log.Println("Server started on port " + appPortStr)
	if err := http.ListenAndServe(":"+appPortStr, nil); err != nil {
		panic(err.Error())
	}
}
