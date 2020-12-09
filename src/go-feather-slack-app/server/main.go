/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 09.12.2020
 * Last Modified Date: 09.12.2020
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package server

import (
	"errors"
	"net/http"
	"os"
	"strconv"
)

type Server struct {
	port int
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func New(listenPort int) *Server {
	return &Server{port: listenPort}
}

func Listen() {
	appPortStr := os.Getenv("APP_PORT")
	if appPortStr == "" {
		panic(errors.New("Environment variable APP_PORT is missing").Error())
	}
	appPort, err := strconv.Atoi(appPortStr)
	if err != nil {
		panic(err.Error())
	}
	New(appPort)
	http.HandleFunc("/healthz", healthz)

	if err := http.ListenAndServe(":"+appPortStr, nil); err != nil {
		panic(err.Error())
	}
}
