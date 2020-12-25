/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 09.12.2020
 * Last Modified Date: 25.12.2020
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

	PodManager "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/manager"
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
		fmt.Fprintf(w, "Job %s has been created on %s with image : %s", FormValues.JobName, FormValues.Namespace, FormValues.DockerImage)

		if err := self.manager.CreateJob(FormValues.Namespace, prefixName, *jobSpec); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error during creation of Job : %s", err.Error())
		}
	}
}

func New(listenPort int, podManager PodManager.PodManager) *Server {
	return &Server{port: listenPort, manager: podManager}
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
	http.HandleFunc("/pods", server.Pods())
	http.HandleFunc("/migrate", server.CreateJob())
	http.HandleFunc("/healthz", healthz)

	log.Println("Server started on port " + appPortStr)
	if err := http.ListenAndServe(":"+appPortStr, nil); err != nil {
		panic(err.Error())
	}
}
