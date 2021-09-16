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
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
	"github.com/sirupsen/logrus"
)

func AddRoutes(router *mux.Router, svc *Service, logger logrus.FieldLogger) {
	router.HandleFunc("/api/v1/teams/create/submit", handleCreateTeam(svc, logger))
	router.HandleFunc("/api/v1/teams/list/submit", handleGetTeams(svc, logger))
	router.HandleFunc("/api/v1/gamedays/create/lookup", handleGamedayLookupTeams(svc, logger))
	router.HandleFunc("/api/v1/gamedays/create/submit", handleCreateGameday(svc, logger))
	router.HandleFunc("/api/v1/gamedays/list/submit", handleListGameDays(svc, logger))
	router.HandleFunc("/api/v1/gamedays/start/submit", handleStartGameDay(svc, logger))
	router.HandleFunc("/api/v1/gamedays/start/lookup", handleLookupGamedays(svc, logger))
	router.HandleFunc("/api/v1/gamedays/complete/submit", handleCompleteGameDay(svc, logger))
	router.HandleFunc("/api/v1/gamedays/complete/lookup", handleLookupGamedays(svc, logger))
	router.HandleFunc("/api/v1/gamedays/cancel/submit", handleCancelGameDay(svc, logger))
	router.HandleFunc("/api/v1/gamedays/cancel/lookup", handleLookupGamedays(svc, logger))
}

func handleCreateTeam(svc *Service, logger logrus.FieldLogger) http.HandlerFunc {
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
			logger.WithError(err).Error("failed to unmarshal json")
			transport.WriteBadRequestError(w, err)
			return
		}
		if err := dto.Validate(); err != nil {
			logger.WithError(err).Error("failed to validate request")
			transport.WriteBadRequestError(w, err)
			return
		}

		members, err := svc.CreateTeam(dto)
		if err != nil {
			logger.WithError(err).Error("failed to create team")
			transport.WriteBadRequestError(w, err)
			return
		}
		msg := fmt.Sprintf("You are added in Team: **%s**", strings.ToUpper(dto.Name))
		mmclient.AsBot(call.Context).DM(dto.Member.UserID, msg)

		transport.WriteJSON(w, apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: getMarkdown(members),
		})
	}
}

func handleGetTeams(svc *Service, logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teams, err := svc.GetTeams()
		if err != nil {
			logger.WithError(err).Error("failed to get teams")
			transport.WriteBadRequestError(w, err)
			return
		}

		transport.WriteJSON(w, apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: getMarkdown(teams),
		})
	}
}

func handleGamedayLookupTeams(svc *Service, logger logrus.FieldLogger) http.HandlerFunc {
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

func handleCreateGameday(svc *Service, logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		call, err := apps.CallRequestFromJSONReader(r.Body)
		if err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}

		jsonString, err := json.Marshal(call.Values)
		if err != nil {
			logger.WithError(err).Error("failed to unmarshal create gameday request")
			transport.WriteBadRequestError(w, err)
			return
		}
		var dto GamedayDTO
		if err := json.Unmarshal(jsonString, &dto); err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}
		if err := dto.Validate(); err != nil {
			logger.WithError(err).Error("failed to validate request")
			transport.WriteBadRequestError(w, err)
			return
		}
		if err := svc.CreateGameday(call.Context, dto); err != nil {
			logger.WithError(err).Error("failed to create gameday")
			transport.WriteBadRequestError(w, err)
			return
		}

		transport.WriteJSON(w, apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: md.MD(fmt.Sprintf("Gameday **%s** scheduled succesfully", dto.Name)),
		})
	}
}

func handleListGameDays(svc *Service, logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gamedays, err := svc.ListGamedays()
		if err != nil {
			logger.WithError(err).Error("failed to list gamedays")
			transport.WriteBadRequestError(w, err)
			return
		}

		transport.WriteJSON(w, apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: getGameDaysMarkdown(gamedays),
		})
	}
}

func handleLookupGamedays(svc *Service, logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		call, err := apps.CallRequestFromJSONReader(r.Body)
		if err != nil {
			transport.WriteBadRequestError(w, err)
			return
		}
		if call.SelectedField != "id" {
			transport.WriteBadRequestError(w, fmt.Errorf("unexpected lookup field: %s", call.SelectedField))
			return
		}

		var states []string
		if strings.Contains(call.Path, "start") {
			states = append(states, string(GamedayScheduledState))
		} else if strings.Contains(call.Path, "complete") {
			states = append(states, string(GamedayInProgressState))
		} else {
			states = append(states, string(GamedayScheduledState), string(GamedayInProgressState))
		}
		gamedays, err := svc.LookupGamedays(states)
		if err != nil {
			logger.WithField("states", states).WithError(err).Error("failed to lookup gamedays by state")
			transport.WriteBadRequestError(w, err)
			return
		}
		transport.WriteJSON(w, apps.CallResponse{
			Type: apps.CallResponseTypeOK,
			Data: map[string]interface{}{
				"items": gamedays,
			},
		})
	}
}

func handleStartGameDay(svc *Service, logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto, err := parseUpdateGamedayStateDto(r)
		if err != nil {
			logger.WithError(err).Error("failed to parse gameday state")
			transport.WriteBadRequestError(w, err)
			return
		}
		if err := svc.UpdateGamedayState(dto.Value, GamedayInProgressState); err != nil {
			logger.WithField("ID", dto.Value).WithError(err).Error("failed to start the gameday")
			transport.WriteBadRequestError(w, err)
			return
		}
		transport.WriteJSON(w, apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: md.MD("Gameday just started"),
		})
	}
}

func handleCompleteGameDay(svc *Service, logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto, err := parseUpdateGamedayStateDto(r)
		if err != nil {
			logger.WithError(err).Error("failed to parse gameday state")
			transport.WriteBadRequestError(w, err)
			return
		}
		if err := svc.UpdateGamedayState(dto.Value, GamedayCompletedState); err != nil {
			logger.WithField("ID", dto.Value).WithError(err).Error("failed to complete the gameday")
			transport.WriteBadRequestError(w, err)
			return
		}
		transport.WriteJSON(w, apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: md.MD("Gameday just completed"),
		})
	}
}
func handleCancelGameDay(svc *Service, logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto, err := parseUpdateGamedayStateDto(r)
		if err != nil {
			logger.WithError(err).Error("failed to parse gameday state")
			transport.WriteBadRequestError(w, err)
			return
		}
		if err := svc.UpdateGamedayState(dto.Value, GamedayCancelledState); err != nil {
			logger.WithField("ID", dto.Value).WithError(err).Error("failed to cancel the gameday")
			transport.WriteBadRequestError(w, err)
			return
		}
		transport.WriteJSON(w, apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: md.MD("Gameday just cancelled"),
		})
	}
}

func parseUpdateGamedayStateDto(r *http.Request) (LookupDTO, error) {
	call, err := apps.CallRequestFromJSONReader(r.Body)
	if err != nil {
		return LookupDTO{}, err
	}

	jsonString, err := json.Marshal(call.Values)
	if err != nil {
		return LookupDTO{}, err
	}
	var dto UpdateGameDayStateDTO
	if err := json.Unmarshal(jsonString, &dto); err != nil {
		return LookupDTO{}, err
	}
	return dto.ID, nil

}
