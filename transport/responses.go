package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func WriteBadRequestError(w http.ResponseWriter, err error) {
	WriteJSON(w, newCallErrorResponse(fmt.Sprintf("Invalid request. Error: %s", err.Error())))
}

func newCallErrorResponse(message string) apps.CallResponse {
	return apps.CallResponse{
		Type:      apps.CallResponseTypeError,
		ErrorText: message,
	}
}

func WriteJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
