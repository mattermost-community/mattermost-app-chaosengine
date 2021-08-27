package gameday

import (
	"fmt"
	"strings"

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

func (t Team) toLookupTeamDTO() LookupTeamDTO {
	return LookupTeamDTO{
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

// Gameday describes the team and the member included
// on this gameday. Different teams can set different
// gamedays
type Gameday struct {
	ID          string `db:"id"`
	Title       string `db:"title"`
	TeamID      string `db:"team_id"`
	ScheduledAt int64  `db:"scheduled_at"`
	CreatedAt   int64  `db:"created_at"`
	UpdatedAt   int64  `db:"updated_at"`
}

// getMarkdown for team members
func getMarkdown(members []TeamMember) md.MD {
	txt := "| team | members | \n"
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
