package gameday

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-app-chaosengine/transport"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/sirupsen/logrus"
)

func AddRoutes(router *mux.Router, svc *Service, logger logrus.FieldLogger) {
	router.HandleFunc("/api/v1/teams/create/submit", handleCreateTeam(svc))
	router.HandleFunc("/api/v1/teams/list/submit", handleGetTeams(svc))
	router.HandleFunc("/api/v1/gamedays/create/lookup", handleGamedayLookupTeams(svc))
	router.HandleFunc("/api/v1/gamedays/create/submit", nil) //TBD
	router.HandleFunc("/api/v1/gamedays/list/form", nil)     //TBD
	router.HandleFunc("/api/v1/gamedays/start/submit", nil)  //TBD
}

func handleCreateTeam(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		call, err := apps.CallRequestFromJSONReader(r.Body)
		if err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}

		jsonString, err := json.Marshal(call.Values)
		if err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}
		var dto CreateTeamDTO
		if err := json.Unmarshal(jsonString, &dto); err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}
		if err := dto.Validate(); err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}

		members, err := svc.CreateTeam(dto)
		if err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}
		mmclient.AsBot(call.Context).
			DM(dto.Member.UserID, "You are added in Team: **%s**", strings.ToUpper(dto.Name))

		transport.WriteJSON(w, apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: getMarkdown(members),
		})
	}
}

func handleGetTeams(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teams, err := svc.GetTeams()
		if err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}

		transport.WriteJSON(w, apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: getMarkdown(teams),
		})
	}
}

func handleGamedayLookupTeams(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		call, err := apps.CallRequestFromJSONReader(r.Body)
		if err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}
		if call.SelectedField != "team" {
			transport.WriteBadRequestError(w, fmt.Errorf("unexpected lookup field: %s", call.SelectedField))
			return
		}

		teams, err := svc.LookupTeams()
		if err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}
		transport.WriteJSON(w, apps.CallResponse{
			Type: apps.CallResponseTypeOK,
			Data: map[string]interface{}{
				"items": teams,
			},
		})
	}
}
