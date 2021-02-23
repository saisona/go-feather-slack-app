/**
 * File              : utils.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 04.01.2021
 * Last Modified Date: 18.02.2021
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
	"regexp"

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

//Allow fetch http request Body to a know specified structure
//@args httpBody: The http.Request.Body in a http.HandlerFunc
//@args structHandler: The wanted structure to hold the data.
//@returns: error if any occurs, nil otherwise.
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

//Send Slack message as an API endpoint response.
//@used: with slach commands
//@args message: is the message to send
//@args w: is the http.ResponseWriter to use to send the message
func SendSlackMessage(message string, w http.ResponseWriter) {
	params := &slack.Msg{Text: message}
	b, err := json.Marshal(params)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR : %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(b); err != nil {
		log.Printf("Error during sending slack message = %s", err.Error())
		return
	}
	return
}

// Send Slack message using the API call
//@used: used to send async message
//@args message:, is the message to send
//@args threadTs:, is the message to send
//@returns: (string, error) where string is the thread_ts.
func (self *Server) sendSlackMessageWithClient(message string, threadTs string) (string, error) {
	var Options slack.MsgOption

	OptionMessage := slack.MsgOptionText(message, false)

	if threadTs != "" {
		OptionTs := slack.MsgOptionTS(threadTs)
		Options = slack.MsgOptionCompose(OptionTs, OptionMessage)
	} else {
		Options = slack.MsgOptionCompose(OptionMessage)
	}

	_, thread_ts, err := self.slackClient.PostMessage(self.config.SLACK_ANSWER_CHANNEL_ID, Options)
	if err != nil {
		return "", err
	}
	return thread_ts, nil
}

func generateDefaultAnswerMention() string {
	possibleAnswers := []string{"Hello there !", "What can I do for you!", "Work work work everyday, everyday the same work!", "Oh I hope this time it'll work!", "When can I'll take a break?"}
	indexAnswer := rand.Intn(5)
	return possibleAnswers[indexAnswer]
}

func (self *Server) isValidVersion(payload string) bool {
	version := payload
	versionRegex, _ := regexp.Compile("v[0-9]+\\.[0-9]+\\.[0-9]+")
	return versionRegex.MatchString(version)
}

func initConfig() *ServerConfig {

	SLACK_API_TOKEN := os.Getenv("SLACK_API_TOKEN")
	SLACK_SIGNING_SECRET := os.Getenv("SLACK_SIGNING_SECRET")
	SLACK_ANSWER_CHANNEL_ID := os.Getenv("SLACK_ANSWER_CHANNEL_ID")

	DOCKER_IMAGE := os.Getenv("APP_DOCKER_IMAGE")
	MIGRATION_COMMAND := os.Getenv("APP_MIGRATION_COMMAND")
	SEED_COMMAND := os.Getenv("APP_SEED_COMMAND")
	SEQUELIZE_MIGRATION_ENV_NAME := os.Getenv("APP_SEQUELIZE_MIGRATION_ENV_NAME")
	SEQUELIZE_SEED_ENV_NAME := os.Getenv("APP_SEQUELIZE_SEED_ENV_NAME")

	if SLACK_API_TOKEN == "" || DOCKER_IMAGE == "" || SEQUELIZE_MIGRATION_ENV_NAME == "" || SEQUELIZE_SEED_ENV_NAME == "" || SLACK_SIGNING_SECRET == "" || SLACK_ANSWER_CHANNEL_ID == "" {
		log.Panicln(errors.New("One of [APP_DOCKER_IMAGE, SLACK_API_TOKEN, APP_SEQUELIZE_SEED_ENV_NAME, APP_SEQUELIZE_MIGRATION_ENV_NAME, SLACK_SIGNING_SECRET, SLACK_ANSWER_CHANNEL_ID] environment variables is missing").Error())
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
		SLACK_SIGNING_SECRET:         SLACK_SIGNING_SECRET,
		SLACK_ANSWER_CHANNEL_ID:      SLACK_ANSWER_CHANNEL_ID,
		DOCKER_IMAGE:                 DOCKER_IMAGE,
		MIGRATION_COMMAND:            MIGRATION_COMMAND,
		SEED_COMMAND:                 SEED_COMMAND,
		SEQUELIZE_MIGRATION_ENV_NAME: SEQUELIZE_MIGRATION_ENV_NAME,
		SEQUELIZE_SEED_ENV_NAME:      SEQUELIZE_SEED_ENV_NAME,
	}
}
