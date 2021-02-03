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
	log.Printf("#handleSlackChallenge apiEventType=%s body=%s", apiEventType, string(body))
	var payload SlackApiEventPayload

	err := json.Unmarshal(body, &payload)
	if err != nil {
		log.Println("ERROR => ", err.Error())

	}

	log.Printf("Payload => %+v", payload)

	if payload.Type == "url_verification" || apiEventType == slackevents.URLVerification {
		log.Printf("#handleSlackChallenge enter URLVerification Challenge : %s", payload.Challenge)
		challengeResponse := &slackevents.ChallengeResponse{Challenge: payload.Challenge}

		bytes, err := json.Marshal(challengeResponse)

		if err != nil {
			log.Println("Error when Marshall => ", err.Error())
			sendStatusInternalError(w)
		}

		log.Println("Marchalled challenge = ", bytes)

		w.Header().Set("Content-Type", "application/json")
		sent, err := w.Write(bytes)

		if err != nil {
			log.Println("Error when sending response=> ", err.Error())
			sendStatusInternalError(w)
		}
		log.Printf("Sent %d bytes !", sent)
		return
	}
}

func (self *Server) handleSlackEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		api := slack.New(self.config.SLACK_API_TOKEN)
		log.Printf("#handleSlackEvent begin function")

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sendStatusInternalError(w)
		}

		log.Printf("#handleSlackEvent apiEvents creation")
		apiEvents, err := self.initApiEventStruct(body, r)

		if err != nil {
			log.Printf("Error when init ApiEvent = %s", err.Error())
			sendStatusInternalError(w)
			return
		}

		log.Printf("#handleSlackEvent apiEvents created => %+v", apiEvents)
		handleSlackChallenge(apiEvents.Type, body, w, r)

		if apiEvents.Type == slackevents.CallbackEvent {
			log.Println("Entered in CALLBACKEVENT !!! YEAAAH !!")
			innerEvent := apiEvents.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				api.PostMessage(ev.Channel, slack.MsgOptionText(generateDefaultAnswerMention(), false))
			}
		}
	}
}
