package gameday

import (
	"strings"

	"github.com/pkg/errors"
)

// Service respresents the struct for the business logic
// for Gameday service
type Service struct {
	repo GamedayRepository
}

func NewService(repo GamedayRepository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateTeam responsible to create a team by the DTO sent from
// mattermost app
func (s *Service) CreateTeam(dto CreateTeamDTO) ([]TeamMember, error) {
	var teamID string

	team, err := s.repo.GetTeam(strings.ToLower(dto.Name))
	if err != nil {
		return []TeamMember{}, errors.Wrap(err, "failed to get a team in repository")
	}
	if team != nil {
		teamID = team.ID
	} else {
		teamID, err = s.repo.CreateTeam(dto.Name)
		if err != nil {
			return []TeamMember{}, errors.Wrap(err, "failed to create a team in repository")
		}
	}

	if err := s.repo.CreateMember(teamID, dto.Member.UserID, dto.Member.Label); err != nil {
		return []TeamMember{}, errors.Wrap(err, "failed to create a member in repository")
	}
	members, err := s.repo.ListTeams(teamID)
	if err != nil {
		return []TeamMember{}, errors.Wrap(err, "failed to fetch team members in repository")
	}
	return members, nil
}

// LookupTeams responsible to return the teams with a formatted data structure
// so the application can show up the values correctly
func (s *Service) LookupTeams() ([]LookupTeamDTO, error) {
	teams, err := s.repo.GetTeams()
	if err != nil {
		return []LookupTeamDTO{}, errors.Wrap(err, "failed to get teams in repository")
	}
	var results []LookupTeamDTO
	for _, team := range teams {
		results = append(results, team.toLookupTeamDTO())
	}
	return results, nil
}

// GetTeams responsible to return the full list of teams
func (s *Service) GetTeams() ([]TeamMember, error) {
	teams, err := s.repo.GetTeams()
	if err != nil {
		return []TeamMember{}, errors.Wrap(err, "failed to get teams in repository")
	}
	return teams, nil
}
