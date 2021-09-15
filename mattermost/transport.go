package mattermost

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-app-chaosengine/transport"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/pkg/errors"
)

var ErrUnexpectedSignMethod = errors.New("unexpected signing method")
var ErrMissingHeader = errors.Errorf("missing %s: Bearer header", apps.OutgoingAuthHeader)
var ErrActingUserMismatch = errors.New("JWT claim doesn't match actingUserID in context")

type requestHandler func(http.ResponseWriter, *http.Request, *apps.CallRequest)

func AddRoutes(router *mux.Router, m *apps.Manifest, staticAssets fs.FS, localMode bool) {
	router.HandleFunc("/manifest", handleManifest(m))
	router.HandleFunc("/bindings", decodeRequest(handleBindings, localMode))
	router.PathPrefix("/static").Handler(http.StripPrefix("/", http.FileServer(http.FS(staticAssets))))
}

func handleManifest(m *apps.Manifest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transport.WriteJSON(w, m)
	}
}

func handleBindings(w http.ResponseWriter, r *http.Request, c *apps.CallRequest) {
	baseCommand := &apps.Binding{
		Label:       "chaos-engine",
		Icon:        "icon.png",
		Description: "Chaos engine will help teams to run Chaos Gamedays",
		Hint:        "[gameday team]",
	}

	gamedayCommand := &apps.Binding{
		Location:    "gameday",
		Label:       "gameday",
		Icon:        "icon.png",
		Description: "Create and list GameDays",
		Hint:        "[create list start]",
		Bindings: []*apps.Binding{
			{
				Location: "create",
				Label:    "create",
				Form: &apps.Form{
					Fields: []*apps.Field{
						{
							Type:       "text",
							Name:       "name",
							Label:      "name",
							IsRequired: true,
						},
						{
							Type:       "dynamic_select",
							Name:       "team",
							Label:      "team",
							IsRequired: true,
						},
						{
							Type:        "text",
							Name:        "schedule_at",
							Label:       "schedule_at",
							Description: "Format [YYYY-DD-MM HH:MM:SS]",
							IsRequired:  true,
						},
					},
				},
				Call: &apps.Call{
					Path: "/api/v1/gamedays/create",
				},
			},
			{
				Location: "list",
				Label:    "list",
				Form:     &apps.Form{},
				Call: &apps.Call{
					Path: "/api/v1/gamedays/list",
				},
			},
			{
				Location: "start",
				Label:    "start",
				Form: &apps.Form{
					Fields: []*apps.Field{
						{
							Type:       "dynamic_select",
							Name:       "id",
							Label:      "id",
							IsRequired: true,
						},
					},
				},
				Call: &apps.Call{
					Path: "/api/v1/gamedays/start",
				},
			},
			{
				Location: "complete",
				Label:    "complete",
				Form: &apps.Form{
					Fields: []*apps.Field{
						{
							Type:       "dynamic_select",
							Name:       "id",
							Label:      "id",
							IsRequired: true,
						},
					},
				},
				Call: &apps.Call{
					Path: "/api/v1/gamedays/complete",
				},
			},
			{
				Location: "cancel",
				Label:    "cancel",
				Form: &apps.Form{
					Fields: []*apps.Field{
						{
							Type:       "dynamic_select",
							Name:       "id",
							Label:      "id",
							IsRequired: true,
						},
					},
				},
				Call: &apps.Call{
					Path: "/api/v1/gamedays/cancel",
				},
			},
		},
	}
	teamCommand := &apps.Binding{
		Location:    "team",
		Label:       "team",
		Icon:        "icon.png",
		Description: "Create and list teams",
		Hint:        "[create list]",
		Bindings: []*apps.Binding{
			{
				Location: "create",
				Label:    "create",
				Form: &apps.Form{
					Fields: []*apps.Field{
						{
							Type:       "text",
							Name:       "name",
							Label:      "name",
							IsRequired: true,
						},
						{
							Type:       "user",
							Name:       "member",
							Label:      "member",
							IsRequired: true,
						},
					},
				},
				Call: &apps.Call{
					Path: "/api/v1/teams/create",
				},
			}, {
				Location: "list",
				Label:    "list",
				Form:     &apps.Form{},
				Call: &apps.Call{
					Path: "/api/v1/teams/list",
				},
			},
		},
	}

	baseCommand.Bindings = append(baseCommand.Bindings, gamedayCommand)
	baseCommand.Bindings = append(baseCommand.Bindings, teamCommand)

	commands := &apps.Binding{
		Location: apps.LocationCommand,
		Bindings: []*apps.Binding{
			baseCommand,
		},
	}

	call := &apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: []*apps.Binding{
			commands,
		},
	}
	transport.WriteJSON(w, call)
}

func decodeRequest(f requestHandler, localMode bool) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		data, err := apps.CallRequestFromJSONReader(r.Body)
		if err != nil {
			transport.WriteBadRequestError(rw, err)
			return
		}

		if localMode {
			claims, err := checkJWT(r)
			if err != nil {
				transport.WriteBadRequestError(rw, err)
				return
			}

			if data.Context.ActingUserID != "" && data.Context.ActingUserID != claims.ActingUserID {
				transport.WriteBadRequestError(rw, ErrActingUserMismatch)
				return
			}
		}

		f(rw, r, nil)
	}
}
func checkJWT(req *http.Request) (*apps.JWTClaims, error) {
	authValue := req.Header.Get(apps.OutgoingAuthHeader)
	if !strings.HasPrefix(authValue, "Bearer ") {
		return nil, ErrMissingHeader
	}

	jwtoken := strings.TrimPrefix(authValue, "Bearer ")
	claims := apps.JWTClaims{}
	_, err := jwt.ParseWithClaims(jwtoken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSignMethod, token.Header["alg"])
		}
		return []byte("1234"), nil
	})

	if err != nil {
		return nil, err
	}

	return &claims, nil
}
