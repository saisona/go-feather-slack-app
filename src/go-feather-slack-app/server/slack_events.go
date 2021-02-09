/**
 * File              : slack_events.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 23.01.2021
 * Last Modified Date: 09.02.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

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
				textMessage := generateDefaultAnswerMention()
				threadTs, err := self.sendSlackMessageWithClient(textMessage, "")
				if err != nil {
					log.Printf("Error when posting message on slack thread_ts=%s and err=%s", threadTs, err.Error())
				}
			default:
				log.Printf("Enter default = %+v", ev)
			}

		}
	}
}
