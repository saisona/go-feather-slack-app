/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 09.12.2020
 * Last Modified Date: 16.12.2020
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

func (self *Server) Pods() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write(nil)
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
		json.Marshal(pods.Items)
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
	http.HandleFunc("/healthz", healthz)

	if err := http.ListenAndServe(":"+appPortStr, nil); err != nil {
		panic(err.Error())
	}
}
