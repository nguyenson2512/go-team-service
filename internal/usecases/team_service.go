package usecases

import (
	"fmt"
	"team-service/internal/entities"
	"team-service/internal/repository"
)

type TeamService interface {
	CreateTeam(teamName string, managers []entities.Manager, members []entities.Member) (map[string]interface{}, error)
	AddMember(teamID uint, memberID string) error
	DeleteMember(teamID uint, memberID string) error
	AddManager(teamID uint, managerID string) error
	DeleteManager(teamID uint, managerID string) error
}

type teamService struct {
	teamRepo repository.TeamRepository
}

func NewTeamService(teamRepo repository.TeamRepository) TeamService {
	return &teamService{
		teamRepo: teamRepo,
	}
}

func (s *teamService) CreateTeam(teamName string, managers []entities.Manager, members []entities.Member) (map[string]interface{}, error) {
	team := &entities.Team{
		TeamName: teamName,
	}

	err := s.teamRepo.Create(team)
	if err != nil {
		return nil, err
	}

	// Add managers to roster
	for _, m := range managers {
		roster := &entities.Roster{
			TeamId:   team.TeamId,
			UserId:   m.ManagerId,
			IsLeader: true,
		}
		s.teamRepo.CreateRoster(roster)
	}

	// Add members to roster
	for _, m := range members {
		roster := &entities.Roster{
			TeamId:   team.TeamId,
			UserId:   m.MemberId,
			IsLeader: false,
		}
		s.teamRepo.CreateRoster(roster)
	}

	return map[string]interface{}{
		"teamId":   team.TeamId,
		"teamName": team.TeamName,
		"managers": managers,
		"members":  members,
	}, nil
}

func (s *teamService) AddMember(teamID uint, memberID string) error {
	roster := &entities.Roster{
		TeamId:   teamID,
		UserId:   memberID,
		IsLeader: false,
	}
	return s.teamRepo.CreateRoster(roster)
}

func (s *teamService) DeleteMember(teamID uint, memberID string) error {
	return s.teamRepo.DeleteRoster(teamID, memberID, false)
}

func (s *teamService) AddManager(teamID uint, managerID string) error {
	roster := &entities.Roster{
		TeamId:   teamID,
		UserId:   managerID,
		IsLeader: true,
	}
	return s.teamRepo.CreateRoster(roster)
}

func (s *teamService) DeleteManager(teamID uint, managerID string) error {
	return s.teamRepo.DeleteRoster(teamID, managerID, true)
}

// Helper function to parse string to uint
func parseUint(s string) uint {
	var v uint
	fmt.Sscanf(s, "%d", &v)
	return v
}
