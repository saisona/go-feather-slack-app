/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 09.12.2020
 * Last Modified Date: 18.02.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package server

import (
	"errors"
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

func (self *Server) fromSlackTextToStruct(commandName string, slackTextArguments []string, structHandler *JobCreationPayload) error {
	dockerTag := slackTextArguments[0]

	structHandler.DockerImage = self.config.DOCKER_IMAGE + ":" + dockerTag
	structHandler.Namespace = "default"
	structHandler.JobName = "go-feather-slack-app-" + strconv.Itoa(int(time.Now().Unix()))
	structHandler.EnvVariablesMap = make(map[string]string)

	if commandName == self.config.MIGRATION_COMMAND {
		structHandler.EnvVariablesMap[self.config.SEQUELIZE_MIGRATION_ENV_NAME] = slackTextArguments[1]
		structHandler.ConfigMapsNames = slackTextArguments[2:]
	} else if commandName == self.config.SEED_COMMAND {
		structHandler.EnvVariablesMap[self.config.SEQUELIZE_SEED_ENV_NAME] = slackTextArguments[1]
		structHandler.ConfigMapsNames = slackTextArguments[2:]
	} else {
		return errors.New("Neither migration nor seed command has been provided")
	}

	return nil
}

func (self *Server) SubmitJobCreation(commandName string, slackTextArguments []string, w http.ResponseWriter, r *http.Request) {
	var FormValues JobCreationPayload

	if err := self.fromSlackTextToStruct(commandName, slackTextArguments, &FormValues); err != nil {
		SendSlackMessage("An error occured while unmarchalling your payload : "+err.Error(), w)
	}

	configMapRefs := self.manager.CreateConfigRefSpec(FormValues.ConfigMapsNames)
	envMapRefs := self.manager.CreateEnvsRefSpec(FormValues.EnvVariablesMap)
	prefixName := FormValues.JobName + "-job"
	jobSpec := self.manager.CreateJobSpec("go-feather-slack-app-job", prefixName, FormValues.DockerImage, envMapRefs, configMapRefs)
	pod, err := self.manager.CreateJob(FormValues.Namespace, prefixName, *jobSpec)
	threadTs, err := self.sendSlackMessageWithClient("Creation of job "+pod.Name, "")

	if err != nil {
		log.Printf("Error during creation of Job: %s", err.Error())
		SendSlackMessage("Error during creation of Job: "+err.Error(), w)
		return
	}

	self.sendSlackMessageWithClient("Job has been created, I'll send logs when finished", threadTs)
	self.sendSlackMessageWithClient("Image :"+FormValues.DockerImage, threadTs)
	self.FetchJobPodLogs(FormValues.Namespace, pod.Name, threadTs)
	SendSlackMessage("Job has been created", w)
}

func (self *Server) FetchJobPodLogs(podNamespace string, podName string, threadTs string) {
	log.Printf("[BEFORE] GetPodLogs podNamespace=%s podName=%s", podNamespace, podName)
	logs, podStatus, err := self.manager.GetPodLogs(podNamespace, podName)
	log.Printf("podStatus = %s", podStatus)

	if err != nil {
		self.sendSlackMessageWithClient(err.Error(), "")
		return
	}

	log.Printf("Sending back logs to slack channel")
	self.sendSlackMessageWithClient("Job "+podName+" "+podStatus, threadTs)
	self.sendSlackMessageWithClient(logs, threadTs)
}

func (self *Server) handleSlackCommand() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, self.config.SLACK_SIGNING_SECRET)
		if err != nil {
			log.Println("Error creating NewSecretVerifier: ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			log.Println("Error parsing SlachCommandParse: ", err.Error())
			sendStatusInternalError(w)
			return
		}

		switch s.Command {
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
		case self.config.MIGRATION_COMMAND:
			slackTextArguments := strings.Fields(s.Text)
			err := r.ParseForm()
			if err != nil {
				SendSlackMessage("Error : "+err.Error(), w)
				log.Println("[ERROR] " + err.Error())
				return
			}

			if len(slackTextArguments) < 2 {
				SendSlackMessage("You must at least specify a version and a migration name!", w)
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
	slackClient := slack.New(appConfig.SLACK_API_TOKEN)
	return &Server{port: listenPort, manager: podManager, config: *appConfig, slackClient: *slackClient}
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
	http.HandleFunc("/events", server.handleSlackEvent())
	http.HandleFunc("/healthz", healthz)

	log.Println("Server started on port " + appPortStr)
	if err := http.ListenAndServe(":"+appPortStr, nil); err != nil {
		panic(err.Error())
	}
}
