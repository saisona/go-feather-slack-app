/**
 * File              : slack_events.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 23.01.2021
 * Last Modified Date: 27.02.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func (self *Server) initApiEventStruct(body []byte, r *http.Request) (slackevents.EventsAPIEvent, error) {

	verifier := slackevents.TokenComparator{VerificationToken: self.config.SLACK_SIGNING_SECRET}
	apiEvents, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(verifier))

	if err != nil {
		log.Println("Error while creation of ParsedEvent with body", err.Error())
		return slackevents.EventsAPIEvent{}, err
	}

	return apiEvents, nil
}

func handleSlackChallenge(apiEventType string, body []byte, w http.ResponseWriter, r *http.Request) {
	var payload SlackApiEventPayload

	err := json.Unmarshal(body, &payload)
	if err != nil {
		log.Println("ERROR => ", err.Error())

	}

	if payload.Type == "url_verification" || apiEventType == slackevents.URLVerification {
		challengeResponse := &slackevents.ChallengeResponse{Challenge: payload.Challenge}

		bytes, err := json.Marshal(challengeResponse)

		if err != nil {
			log.Println("Error when Marshall => ", err.Error())
			sendStatusInternalError(w)
		}

		w.Header().Set("Content-Type", "application/json")

		if _, err := w.Write(bytes); err != nil {
			log.Println("Error when sending response=> ", err.Error())
			sendStatusInternalError(w)
		}
		return
	}
}

func (self *Server) handleSlackEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sendStatusInternalError(w)
		}

		apiEvents, err := self.initApiEventStruct(body, r)

		if err != nil {
			log.Printf("Error when init ApiEvent = %s", err.Error())
			sendStatusInternalError(w)
			return
		}

		handleSlackChallenge(apiEvents.Type, body, w, r)

		if apiEvents.Type == slackevents.CallbackEvent {

			InnerEventData := apiEvents.InnerEvent.Data
			switch ev := InnerEventData.(type) {
			case *slackevents.AppMentionEvent:
				log.Println("Found the mention")

				// Remove the bot mention id
				userSlackMessage := strings.SplitAfterN(ev.Text, " ", 1)
				userId := ev.User
				messageThreadTS := ev.ThreadTimeStamp
				log.Printf("User %s send %s", userId, userSlackMessage)
				self.handleMention(messageThreadTS, userId, userSlackMessage[0], strings.Join(userSlackMessage[1:], " "))

				//textMessage := generateDefaultAnswerMention()
				//threadTs, err := self.sendSlackMessageWithClient(textMessage, "")
				//if err != nil {
				//	log.Printf("Error when posting message on slack thread_ts=%s and err=%s", threadTs, err.Error())
				//}
			default:
				log.Printf("Enter default = %+v", ev)
			}

		}
	}
}

// handleMention is used to trigger action when mentioning slack bot based on the message.
// @arg userId: Id of the Slack User who did mentioned the bot.
// @arg actionType: substring of received slack message that triggers actions.
// @arg args: rest of the slack message that can be used as action payload.
func (self *Server) handleMention(threadTs string, userId string, actionType string, args string) {
	log.Printf("Received from %s action %s with payload : %s", userId, actionType, args)

	switch actionType {

	case "list-models":
		self.sendSlackMessageWithClient("Available models : ", threadTs)
		//TODO : fetch deployed model through the API.
		//TODO : authenticate user to API
		//TODO : send authenticated request to https://api.inarix.com/imodels/model-template
		//TODO : send authenticated request to https://api.inarix.com/imodels/model-instance
		//TODO : for each model-template.prodModelInstanceId -> fetch name of model-instance.name with the same id
	case ":rocket:":
		//TODO : handle deploy model with api registration
		log.Printf("Deploying model %s", args)
		self.sendSlackMessageWithClient("What kind of model do you want to deploy?", threadTs)
		self.sendModelDeployerPopup(userId)

	default:
		log.Printf("No actionType found to be handled")
	}

}

func mapStrListToOptionBlock(options map[string]string, description string) *[]*slack.OptionBlockObject {
	optionArray := make([]*slack.OptionBlockObject)
	for key, value := range options {
		text := &slack.TextBlockObject{Type: "plain_text", Text: key, Emoji: true}
		desc := &slack.TextBlockObject{Type: "plain_text", Text: description, Emoji: true}
		obj := slack.NewOptionBlockObject(value, text, desc)
		append(optionArray, obj)
	}
	return optionArray
}

// sendModelDeployerPopup will send Slack block to help deploy model using slack
// @arg userId: ID of the user who triggered model deployment
func (self *Server) sendModelDeployerPopup(userId string) {
	//TODO: Create Block type
	//	contextName := &slack.TextBlockObject{Type: "plain_text", Text: "Deploy model on production :rocket:", Emoji: true}
	contextName := slack.NewTextBlockObject("plain_text", "Deploy model on production :rocket:", true, true)
	optionsSelect := mapStrListToOptionBlock(map[string]string{"barley-variety": "mt-barley-variety", "corn": "corn/dent", "Soft Wheat": "soft-wheat"})
	//	staticSelect := slack.NewOptionsSelectBlockElement("static_select", &slack.TextBlockObject{Type: "plain_text"}, "deploy-model", *optionsSelect...)
	//
	//	messageToSend := slack.NewBlockMessage(contextName)
	//	message := slack.AddBlockMessage(messageToSend, staticSelect)
	//	slack.Add
	//	messageOption := slack.MsgOptionBlocks(message, staticSelect)
	//	self.slackClient.SendMessage(self.config.SLACK_ANSWER_CHANNEL_ID, messageOption)

}
