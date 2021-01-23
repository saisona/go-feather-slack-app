/**
 * File              : slack_events.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 23.01.2021
 * Last Modified Date: 23.01.2021
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

	verifier, err := slack.NewSecretsVerifier(r.Header, self.config.SLACK_API_TOKEN)
	if err != nil {
		return slackevents.EventsAPIEvent{}, err
	}

	if err := verifier.Ensure(); err != nil {
		return slackevents.EventsAPIEvent{}, err
	}

	apiEvents, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
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
		log.Printf("#handleSlackChallenge enter URLVerification")
		var response *slackevents.ChallengeResponse
		response.Challenge = payload.Challenge
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(response.Challenge))
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
			sendStatusInternalError(w)
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
