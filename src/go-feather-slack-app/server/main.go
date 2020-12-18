/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 09.12.2020
 * Last Modified Date: 18.12.2020
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	PodManager "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/manager"
	v1 "k8s.io/api/core/v1"
)

type Server struct {
	port    int
	manager PodManager.PodManager
}

type PodRequestPayload struct {
	namespace string
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func sendStatusMEthodNotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write(nil)
}

func (self *Server) Pods() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendStatusMEthodNotAllowed(w)
			return
		}
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		pods, err := self.manager.GetPods(r.Form.Get("namespace"))
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err.Error())
		}
		marchalled_json, err := json.Marshal(pods.Items)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err.Error())
		}
		fmt.Fprintln(w, string(marchalled_json))
	}
}

func (self *Server) CreateJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendStatusMEthodNotAllowed(w)
			return
		}
		env := []v1.EnvVar{v1.EnvVar{Name: "INARIX_API_HOST", Value: "api.inarix.com"}, v1.EnvVar{Name: "INARIX_API_USERNAME", Value: "toto"}, v1.EnvVar{Name: "INARIX_API_PASSWORD", Value: "tata"}}
		jobSpec := self.manager.CreateJobSpecWithEnvVariable("migration_test", "migration-job", "894517829775.dkr.ecr.eu-west-1.amazonaws.com/inarix-api:v1.14.0-staging", env)
		_, err := self.manager.CreateJob("default", "migration", *jobSpec)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error during creation of Job : %s", err.Error())
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Job %s has been created", jobSpec.Template.Name)

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

	if err := http.ListenAndServe(":"+appPortStr, nil); err != nil {
		panic(err.Error())
	}
}
