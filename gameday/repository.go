package gameday

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/mattermost-app-chaosengine/store"
	"github.com/pkg/errors"
)

// table names in database

const gamedayTableName = "gameday"
const teamTableName = "team"
const memberTableName = "team_member"

// Repository stores a gameday
type Repository struct {
	store *store.SQL
}

// GamedayRepository contract interface for gameday
// repository
type GamedayRepository interface {
	ListGamedays() ([]Gameday, error)
	CreateGameday(gameday Gameday) error
	CreateTeam(name string) (string, error)
	CreateMember(teamID, userID, label string) error
	ListTeams(id string) ([]TeamMember, error)
	GetTeam(name string) (*Team, error)
	GetTeams() ([]TeamMember, error)
}

// NewRepository factory method to create repository
func NewRepository(store *store.SQL) *Repository {
	return &Repository{
		store: store,
	}
}

// ListGamedays returns the list of gamedays created in the app
func (r *Repository) ListGamedays() ([]Gameday, error) {
	q := sq.Select().From(gamedayTableName)

	var gamedays []Gameday
	if err := r.store.SelectBuilder(r.store.DB, &gamedays, q); err != nil {
		return []Gameday{}, errors.Wrap(err, "failed to select gamedays")
	}
	return gamedays, nil
}

// CreateGameday craetes a new gameday in database
func (r *Repository) CreateGameday(gameday Gameday) error {
	insertsMap := map[string]interface{}{
		"id":           store.NewID(),
		"title":        gameday.Title,
		"team_id":      gameday.TeamID,
		"scheduled_at": gameday.ScheduledAt,
		"created_at":   time.Now().UnixNano() / int64(time.Millisecond),
		"updated_at":   0,
	}
	_, err := r.store.ExecBuilder(r.store.DB, sq.Insert(gamedayTableName).SetMap(insertsMap))
	if err != nil {
		return errors.Wrap(err, "failed to create gameday")
	}
	return nil
}

// CreateTeam creates a new team which will be assigned to a Gameday
func (r *Repository) CreateTeam(name string) (string, error) {
	id := store.NewID()
	insertsMap := map[string]interface{}{
		"id":         id,
		"name":       name,
		"created_at": time.Now().UnixNano() / int64(time.Millisecond),
		"updated_at": 0,
	}
	_, err := r.store.ExecBuilder(r.store.DB, sq.Insert(teamTableName).SetMap(insertsMap))
	if err != nil {
		return "", errors.Wrap(err, "failed to create a team")
	}
	return id, nil
}

// ListTeams returns the list of gamedays created in the app for the given
// team ID
func (r *Repository) ListTeams(teamID string) ([]TeamMember, error) {
	sql := `SELECT
		team_member.*,
		team.id "team.id",
		team.name "team.name"
	  FROM
	  team_member JOIN team ON team_member.team_id = team.id
	  WHERE team.id = "%s";`

	sql = fmt.Sprintf(sql, teamID)
	var teamMembers []TeamMember
	if err := r.store.DB.Select(&teamMembers, sql); err != nil {
		return []TeamMember{}, errors.Wrap(err, "failed to list team members")
	}
	return teamMembers, nil
}

// GetTeams returns the list of gamedays created in the app
func (r *Repository) GetTeams() ([]TeamMember, error) {
	sql := `SELECT
		team_member.*,
		team.id "team.id",
		team.name "team.name"
	  FROM
	  team_member JOIN team ON team_member.team_id = team.id;`

	var teamMembers []TeamMember
	if err := r.store.DB.Select(&teamMembers, sql); err != nil {
		return []TeamMember{}, errors.Wrap(err, "failed to get team members")
	}
	return teamMembers, nil
}

// GetTeam returns the team based on the name
func (r *Repository) GetTeam(name string) (*Team, error) {
	q := sq.Select("*").From(teamTableName).Where(sq.Eq{"name": name})
	var teams []Team
	if err := r.store.SelectBuilder(r.store.DB, &teams, q); err != nil {
		return nil, errors.Wrap(err, "failed to find a team")
	}
	if len(teams) == 0 {
		return nil, nil
	}
	return &teams[0], nil
}

// CreateMember creates a new member which will be assigned to a Team
func (r *Repository) CreateMember(teamID, userID, label string) error {
	insertsMap := map[string]interface{}{
		"id":         store.NewID(),
		"team_id":    teamID,
		"user_id":    userID,
		"label":      label,
		"created_at": time.Now().UnixNano() / int64(time.Millisecond),
		"updated_at": 0,
	}
	_, err := r.store.ExecBuilder(r.store.DB, sq.Insert(memberTableName).SetMap(insertsMap))
	if err != nil {
		return errors.Wrap(err, "failed to create a team")
	}
	return nil
}
