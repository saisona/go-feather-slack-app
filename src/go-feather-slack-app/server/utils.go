/**
 * File              : utils.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 04.01.2021
 * Last Modified Date: 04.01.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/slack-go/slack"
)

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
