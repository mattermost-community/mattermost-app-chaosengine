package gameday

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-apps/utils/md"
)

// Team describes the team and the members included on
// this gameday
type Team struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

func (t Team) toLookupTeamDTO() LookupDTO {
	return LookupDTO{
		Label: t.Name,
		Value: t.ID,
	}
}

// Member describes the team and the members included on
// this gameday
type TeamMember struct {
	ID        string `db:"id"`
	TeamID    string `db:"team_id"`
	UserID    string `db:"user_id"`
	Label     string `db:"label"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
	Team      `db:"team"`
}

// GamedayState the state of a gameday
type GamedayState string

const (
	// GamedayScheduledState when we schedule a gameday
	GamedayScheduledState GamedayState = "scheduled"
	// GamedayInProgressState when a gameday is in progress
	GamedayInProgressState GamedayState = "in_progress"
	// GamedayCancelledState when a gameday has been cancelled
	GamedayCancelledState GamedayState = "cancelled"
	// GamedayCompletedState when a gameday has been completed
	GamedayCompletedState GamedayState = "completed"
)

// Gameday describes the team and the member included
// on this gameday. Different teams can set different
// gamedays
type Gameday struct {
	ID          string       `db:"id"`
	Title       string       `db:"title"`
	TeamID      string       `db:"team_id"`
	ScheduledAt int64        `db:"scheduled_at"`
	State       GamedayState `db:"state"`
	CreatedAt   int64        `db:"created_at"`
	UpdatedAt   int64        `db:"updated_at"`
	Team        `db:"team"`
}

func (g Gameday) toGameDayDTO() GamedayDTO {
	return GamedayDTO{
		Name: g.Title,
		Team: LookupDTO{
			Label: g.Team.Name,
			Value: g.Team.ID,
		},
		State:       g.State,
		ScheduledAt: ScheduledAtTime(time.Unix(g.ScheduledAt, 0)),
	}
}

// LookupGamedayDTO describes data transfer object for gameday
type LookupGamedayDTO struct {
	Label string
	Value string
}

func (g Gameday) toLookupGamedayDTO() LookupDTO {
	return LookupDTO{
		Label: g.Title,
		Value: g.ID,
	}
}

// GamedayNominee the nominess for gamedays about
// Master of Disaster
// On Call
type GamedayNominee struct {
	ID                 string `db:"id"`
	GamedayID          string `db:"gameday_id"`
	MemberID           string `db:"member_id"`
	UserID             string `db:"user_id"`
	IsMasterOfDisaster bool   `db:"is_mod"`
	IsOnCall           bool   `db:"is_on_call"`
	CreatedAt          int64  `db:"created_at"`
	UpdatedAt          int64  `db:"updated_at"`
	Gameday            `db:"gameday"`
}

// getMarkdown for team members
func getMarkdown(members []TeamMember) md.MD {
	txt := "| Team | Members | \n"
	txt += "| :-- |:-- |\n"

	var team string
	var users []string
	for _, m := range members {
		team = m.Team.Name
		users = append(users, fmt.Sprintf("@%s", m.Label))
	}
	txt += fmt.Sprintf("|%s|%s|\n", team, strings.Join(users[:], ","))
	return md.MD(txt)
}

// getGameDaysMarkdown makrodnw for the game days
func getGameDaysMarkdown(gamedays []GamedayDTO) md.MD {
	txt := "| Title | Team | Scheduled At | State |\n"
	txt += "| :-- |:-- |:-- |\n"

	for _, g := range gamedays {
		txt += fmt.Sprintf("|%s|%s|%s|%s|\n", g.Name, g.Team.Label, g.ScheduledAt.String(), g.State)
	}
	return md.MD(txt)
}
