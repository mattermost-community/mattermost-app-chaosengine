package gameday

import (
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/mattermost-app-chaosengine/store"
	"github.com/pkg/errors"
)

// table names in database

const gamedayTableName = "gameday"
const teamTableName = "team"
const memberTableName = "team_member"
const nomineeTableName = "gameday_nominee"

// Repository stores a gameday
type Repository struct {
	store *store.SQL
}

// GamedayRepository contract interface for gameday
// repository
type GamedayRepository interface {
	ListGamedays() ([]Gameday, error)
	ListGamedaysByState(states []string) ([]Gameday, error)
	CreateGameday(gameday Gameday) (string, error)
	UpdateGamedayState(gamedayID string, state GamedayState) error
	CreateTeam(name string) (string, error)
	CreateMember(teamID, userID, label string) error
	CreateNominee(nominee GamedayNominee) (string, error)
	ListGamedayNominees(gamedayID string) ([]GamedayNominee, error)
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
	sql := `SELECT
		gameday.*,
		team.name "team.name"
	  FROM
	  gameday INNER JOIN team ON gameday.team_id = team.id
	  WHERE gameday.state IN ('scheduled', 'in_progress');`

	var gamedays []Gameday
	if err := r.store.DB.Select(&gamedays, sql); err != nil {
		return []Gameday{}, errors.Wrap(err, "failed to get gamedays")
	}
	return gamedays, nil
}

// ListGamedaysByState returns the list of gamedays created in the app by the provided state
func (r *Repository) ListGamedaysByState(states []string) ([]Gameday, error) {
	sql := `SELECT
		gameday.*,
		team.name "team.name"
	  FROM
	  gameday INNER JOIN team ON gameday.team_id = team.id
	  WHERE gameday.state IN (%s);`
	commaSepState := "'" + strings.Join(states[:], "', '") + "'"
	sql = fmt.Sprintf(sql, commaSepState)

	var gamedays []Gameday
	if err := r.store.DB.Select(&gamedays, sql); err != nil {
		return []Gameday{}, errors.Wrap(err, "failed to get gamedays")
	}
	return gamedays, nil
}

// CreateGameday creates a new gameday in database
func (r *Repository) CreateGameday(gameday Gameday) (string, error) {
	id := store.NewID()
	insertsMap := map[string]interface{}{
		"id":           id,
		"title":        gameday.Title,
		"team_id":      gameday.TeamID,
		"scheduled_at": gameday.ScheduledAt,
		"state":        GamedayScheduledState,
		"created_at":   time.Now().UnixNano() / int64(time.Millisecond),
		"updated_at":   0,
	}
	_, err := r.store.ExecBuilder(r.store.DB, sq.Insert(gamedayTableName).SetMap(insertsMap))
	if err != nil {
		return "", errors.Wrap(err, "failed to create gameday")
	}
	return id, nil
}

// UpdateGamedayState updates a ngameday state
func (r *Repository) UpdateGamedayState(gamedayID string, state GamedayState) error {
	builder := sq.Update(gamedayTableName).Set("state", state).Where("ID = ?", gamedayID)
	if _, err := r.store.ExecBuilder(r.store.DB, builder); err != nil {
		return errors.Wrap(err, "failed to update gameday state")
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
	  team_member INNER JOIN team ON team_member.team_id = team.id
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
	q := sq.Select("*").From(teamTableName).Where("name LIKE ?", fmt.Sprint("%", name, "%"))
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
		return errors.Wrap(err, "failed to create a team member")
	}
	return nil
}

// Updateember updates an existing member
func (r *Repository) CreateNominee(nominee GamedayNominee) (string, error) {
	id := store.NewID()
	_, err := r.store.ExecBuilder(r.store.DB, sq.
		Insert(nomineeTableName).
		SetMap(map[string]interface{}{
			"id":         id,
			"gameday_id": nominee.GamedayID,
			"member_id":  nominee.MemberID,
			"is_mod":     nominee.IsMasterOfDisaster,
			"is_on_call": nominee.IsOnCall,
			"created_at": time.Now().UnixNano() / int64(time.Millisecond),
			"updated_at": 0,
		}))
	if err != nil {
		return "", errors.Wrapf(err, "failed to create nominee for GamedayID: %s and MemberID: %s", nominee.GamedayID, nominee.MemberID)
	}
	return id, nil
}

// ListGamedayNominees returns the list of gameday nominees by provided gameday ID
func (r *Repository) ListGamedayNominees(gamedayID string) ([]GamedayNominee, error) {
	sql := `SELECT
		gameday_nominee.*,
		gameday.id "gameday.id"
	  FROM
	  gameday_nominee INNER JOIN gameday ON gameday_nominee.gameday_id = gameday.id
	  WHERE gameday.id = "%s" AND gameday.state NOT IN ('scheduled', 'in_progress', 'cancelled', 'completed');`
	sql = fmt.Sprintf(sql, gamedayID)

	var nominees []GamedayNominee
	if err := r.store.DB.Select(&nominees, sql); err != nil {
		return []GamedayNominee{}, errors.Wrap(err, "failed to get gameday nominees")
	}
	return nominees, nil
}
