/**
 * File              : utils.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 04.01.2021
 * Last Modified Date: 20.01.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/slack-go/slack"
)

func sendStatusMethodNotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write(nil)
}

func sendStatusInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "An error occured please look at the logs !")
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

func SendSlackMessage(message string, w http.ResponseWriter) {
	params := &slack.Msg{Text: message}
	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR : %s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
