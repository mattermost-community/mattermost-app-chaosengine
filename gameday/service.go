package gameday

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
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
func (s *Service) LookupTeams() ([]LookupDTO, error) {
	teams, err := s.repo.GetTeams()
	if err != nil {
		return []LookupDTO{}, errors.Wrap(err, "failed to get teams in repository")
	}
	var results []LookupDTO
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

// CreateGameday responsible to create a gameday in database
func (s *Service) CreateGameday(ctx *apps.Context, dto GamedayDTO) error {
	gameday := Gameday{
		Title:       dto.Name,
		TeamID:      dto.Team.Value,
		State:       GamedayScheduledState,
		ScheduledAt: dto.ScheduledAt.Unix(),
	}
	gamedayID, err := s.repo.CreateGameday(gameday)
	if err != nil {
		return errors.Wrap(err, "failed to create a gameday")
	}
	members, err := s.repo.ListTeams(dto.Team.Value)
	if err != nil {
		return errors.Wrap(err, "failed to fetch team members in repository")
	}
	msg := fmt.Sprintf("Gameday: **%s** is scheduled for %s", strings.ToUpper(dto.Name), dto.ScheduledAt.String())
	for _, m := range members {
		mmclient.AsBot(ctx).DM(m.UserID, msg)
	}
	currentNominees, err := s.repo.ListGamedayNominees(gamedayID)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch current gameday nominees")
	}
	filteredMembers := filterOutCurrentNominees(currentNominees, members)
	mod := shuffleAndPickMember(filteredMembers)
	if _, err := s.repo.CreateNominee(GamedayNominee{GamedayID: gamedayID, MemberID: mod.ID, IsMasterOfDisaster: true}); err != nil {
		return errors.Wrapf(err, "failed to nominate a team member for MOD for GamedayID: %s and MemberID: %s", gamedayID, mod.ID)
	}
	oncall := shuffleAndPickMember(filteredMembers)
	if _, err := s.repo.CreateNominee(GamedayNominee{GamedayID: gamedayID, MemberID: oncall.ID, IsOnCall: true}); err != nil {
		return errors.Wrapf(err, "failed to nominate a team member for OnCall for GamedayID: %s and MemberID: %s", gamedayID, mod.ID)
	}
	return nil
}

// UpdateGamedayState updates the state of a gameday accordingly
// to the action
func (s *Service) UpdateGamedayState(gamedayID string, state GamedayState) error {
	return s.repo.UpdateGamedayState(gamedayID, state)
}

// ListGamedays responsible to list the scheduled and in progress gamedays
func (s *Service) ListGamedays() ([]GamedayDTO, error) {
	gamedays, err := s.repo.ListGamedays()
	if err != nil {
		return []GamedayDTO{}, errors.Wrap(err, "failed to get gamedays in repository")
	}
	var results []GamedayDTO
	for _, g := range gamedays {
		results = append(results, g.toGameDayDTO())
	}
	return results, nil
}

// LookupGamedays responsible to lookup the scheduled and in progress gamedays
func (s *Service) LookupGamedays() ([]LookupDTO, error) {
	gamedays, err := s.repo.ListGamedays()
	if err != nil {
		return []LookupDTO{}, errors.Wrap(err, "failed to get gamedays in repository")
	}
	var results []LookupDTO
	for _, g := range gamedays {
		results = append(results, g.toLookupGamedayDTO())
	}
	return results, nil
}

// filterOutCurrentNominees returns only the people who weren't of last gameday as any type of nominee
func filterOutCurrentNominees(nominees []GamedayNominee, members []TeamMember) []TeamMember {
	var filteredMembers []TeamMember
	if len(nominees) == 0 {
		return members
	}
	for _, m := range members {
		for _, n := range nominees {
			if m.ID == n.ID && (n.IsMasterOfDisaster || n.IsOnCall) {
				continue
			}
			filteredMembers = append(filteredMembers, m)
		}
	}
	return filteredMembers
}

// shuffleAndPickMember shuffle the members and picks one randomly
func shuffleAndPickMember(members []TeamMember) *TeamMember {
	if len(members) == 0 {
		return nil
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})
	return &members[0]
}
