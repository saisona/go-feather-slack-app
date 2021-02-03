/**
 * File              : slack_events.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 23.01.2021
 * Last Modified Date: 03.02.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

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
			api := slack.New(self.config.SLACK_API_TOKEN)

			switch ev := apiEvents.InnerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				log.Println("Found the mention")
				textMessage := generateDefaultAnswerMention()
				log.Printf("message was supposed to be %s", textMessage)
				something, someelse, err := api.PostMessage(ev.Channel, slack.MsgOptionText("You sent me "+ev.Text, false))
				if err != nil {
					log.Printf("Error when posting message on slack something=%s someelse=%s and err=%s", something, someelse, err.Error())
				}
			default:
				log.Printf("Enter default = %+v", ev)
			}

		}
	}
}
