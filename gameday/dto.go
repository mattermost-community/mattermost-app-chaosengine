package gameday

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ConfigureDTO the data transfer object for
// configure application database
type ConfigureDTO struct {
	Scheme string `json:"scheme"`
	Url    string `json:"url"`
}

// Validate check if the DTO has the required values
func (c ConfigureDTO) Validate() error {
	if c.Scheme == "" {
		return errors.New("failed: missing required field `scheme`")
	}
	if c.Url == "" {
		return errors.New("failed: missing required field `url`")
	}
	return nil
}

// MemberDTO the data transfer object for
// of a member which matches MM user
type MemberDTO struct {
	Label  string
	UserID string `json:"value"`
}

// Validate check if the DTO has the required values
func (c MemberDTO) Validate() error {
	if c.UserID == "" {
		return errors.New("failed: missing required field `user_id`")
	}
	if c.Label == "" {
		return errors.New("failed: missing required field `label`")
	}
	return nil
}

// CreateTeamDTO the data transfer object for
// creating a new team with a member
type CreateTeamDTO struct {
	Name   string
	Member MemberDTO
}

// Validate check if the DTO has the required values
func (c CreateTeamDTO) Validate() error {
	if c.Name == "" {
		return errors.New("failed: missing required field name")
	}
	return c.Member.Validate()
}

// LookupTeamDTO lookup label value data transfer
// object for team values
type LookupDTO struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// GamedayDTO the data transfer object for
// creating a new gameday
type GamedayDTO struct {
	Name        string
	Team        LookupDTO
	ScheduledAt ScheduledAtTime `json:"schedule_at"`
	State       GamedayState
}

// Validate check if the DTO has the required values
func (g GamedayDTO) Validate() error {
	if g.Name == "" {
		return errors.New("failed: missing required field name")
	}
	if g.Team.Value == "" {
		return errors.New("failed: missing required field team ID")
	}
	if g.ScheduledAt.String() == "" {
		return errors.New("failed: missing required field scheduled_at")
	}
	return nil
}

// ScheduledAtTime time for `scheduled_at`
type ScheduledAtTime time.Time

const timeLayout = "2006-01-02 15:04:05"

// UnmarshalJSON Parses the json string in the custom format
func (ct *ScheduledAtTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(timeLayout, s)
	*ct = ScheduledAtTime(nt)
	return
}

// MarshalJSON writes a quoted string in the custom format
func (ct ScheduledAtTime) MarshalJSON() ([]byte, error) {
	return []byte(timeLayout), nil
}

// Unix the unix for the scheduled at time
func (ct *ScheduledAtTime) Unix() int64 {
	return time.Time(*ct).Unix()
}

// String returns the time in the custom format
func (ct *ScheduledAtTime) String() string {
	t := time.Time(*ct)
	return fmt.Sprintf("%q", t.Format(timeLayout))
}

// UpdateGameDayStateDTO the data transfer object for
// to update the state
type UpdateGameDayStateDTO struct {
	ID LookupDTO `json:"id"`
}
