package gameday

import "errors"

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
type LookupTeamDTO struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
