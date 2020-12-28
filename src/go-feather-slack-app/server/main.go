/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 09.12.2020
 * Last Modified Date: 29.12.2020
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
	"strconv"
	"strings"

	PodManager "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/manager"
	"github.com/slack-go/slack"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func fromBodyToStruct(httpBody io.ReadCloser, structHandler interface{}) error {
	body, err := ioutil.ReadAll(httpBody)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &structHandler); err != nil {
		return err
	}
	return nil
}

func sendStatusMethodNotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write(nil)
}

func sendStatusInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "An error occured please look at the logs !")
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

		var FormValues JobCreationPayload
		if err := fromBodyToStruct(r.Body, &FormValues); err != nil {
			log.Printf("An error occured while unmarchalling your payload : %s", err.Error())
			sendStatusInternalError(w)
			return
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

		fmt.Fprintf(w, "Logs for %s :\n%s", FormValues.JobName, logs)
		//fmt.Fprintf(w, "Job %s has been created on %s with image : %s", FormValues.JobName, FormValues.Namespace, FormValues.DockerImage)
	}
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
			params := &slack.Msg{Text: s.Text}
			b, err := json.Marshal(params)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
		case "/migration":
			params := strings.Split(s.Text, " ")
			log.Println("action param => ", params)
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
	if SLACK_API_TOKEN == "" {
		SLACK_API_TOKEN = "0f87b271e98de06b32f1fe5ec7014c07"
	}
	return &Server{port: listenPort, manager: podManager, SLACK_API_TOKEN: SLACK_API_TOKEN}
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
	http.HandleFunc("/migrate", server.CreateJob())
	http.HandleFunc("/healthz", healthz)

	log.Println("Server started on port " + appPortStr)
	if err := http.ListenAndServe(":"+appPortStr, nil); err != nil {
		panic(err.Error())
	}
}
