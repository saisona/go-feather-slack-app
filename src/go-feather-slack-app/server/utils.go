/**
 * File              : utils.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 04.01.2021
 * Last Modified Date: 23.01.2021
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
	"math/rand"
	"net/http"
	"os"

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

func generateDefaultAnswerMention() string {
	possibleAnswers := make([]string, 5)
	possibleAnswers = append(possibleAnswers, "Hello there !")
	possibleAnswers = append(possibleAnswers, "What can I do for you!")
	possibleAnswers = append(possibleAnswers, "Work work work everyday, everyday the same work!")
	possibleAnswers = append(possibleAnswers, "Oh I hope this time it'll work!")
	possibleAnswers = append(possibleAnswers, "When can I'll take a break?")
	indexAnswer := rand.Intn(5)
	return possibleAnswers[indexAnswer]
}

func initConfig() *ServerConfig {

	SLACK_API_TOKEN := os.Getenv("SLACK_API_TOKEN")
	DOCKER_IMAGE := os.Getenv("APP_DOCKER_IMAGE")
	MIGRATION_COMMAND := os.Getenv("APP_MIGRATION_COMMAND")
	SEED_COMMAND := os.Getenv("APP_SEED_COMMAND")
	SEQUELIZE_MIGRATION_ENV_NAME := os.Getenv("APP_SEQUELIZE_MIGRATION_ENV_NAME")
	SEQUELIZE_SEED_ENV_NAME := os.Getenv("APP_SEQUELIZE_SEED_ENV_NAME")

	if SLACK_API_TOKEN == "" || DOCKER_IMAGE == "" || SEQUELIZE_MIGRATION_ENV_NAME == "" || SEQUELIZE_SEED_ENV_NAME == "" {
		log.Panicln(errors.New("One of [APP_DOCKER_IMAGE, SLACK_API_TOKEN, APP_SEQUELIZE_SEED_ENV_NAME, APP_SEQUELIZE_MIGRATION_ENV_NAME] environment variables is missing").Error())
	}

	if MIGRATION_COMMAND == "" {
		log.Println("WARNING: You didn't specified any APP_MIGRATION_COMMAND, default /migration will be used")
		MIGRATION_COMMAND = "/migration"
	}

	if SEED_COMMAND == "" {
		log.Println("WARNING: You didn't specified any APP_SEED_COMMAND, default /seed will be used")
		MIGRATION_COMMAND = "/seed"
	}

	return &ServerConfig{
		SLACK_API_TOKEN:              SLACK_API_TOKEN,
		DOCKER_IMAGE:                 DOCKER_IMAGE,
		MIGRATION_COMMAND:            MIGRATION_COMMAND,
		SEED_COMMAND:                 SEED_COMMAND,
		SEQUELIZE_MIGRATION_ENV_NAME: SEQUELIZE_MIGRATION_ENV_NAME,
		SEQUELIZE_SEED_ENV_NAME:      SEQUELIZE_SEED_ENV_NAME,
	}
}
