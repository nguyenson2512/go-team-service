package usecases

import (
	"fmt"
	"team-service/internal/entities"
	"team-service/internal/kafka"
	"team-service/internal/repository"
	"time"
)

type TeamService interface {
	CreateTeam(teamName string, managers []entities.Manager, members []entities.Member, performedBy string) (map[string]interface{}, error)
	AddMember(teamID uint, memberID string, performedBy string) error
	DeleteMember(teamID uint, memberID string, performedBy string) error
	AddManager(teamID uint, managerID string, performedBy string) error
	DeleteManager(teamID uint, managerID string, performedBy string) error
}

type teamService struct {
	teamRepo      repository.TeamRepository
	kafkaProducer kafka.TeamEventProducer
}

func NewTeamService(teamRepo repository.TeamRepository, kafkaProducer kafka.TeamEventProducer) TeamService {
	return &teamService{
		teamRepo:      teamRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (s *teamService) CreateTeam(teamName string, managers []entities.Manager, members []entities.Member, performedBy string) (map[string]interface{}, error) {
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

	// Send Kafka event for team creation
	if s.kafkaProducer != nil {
		event := kafka.TeamEvent{
			EventType:   kafka.TeamCreated,
			TeamId:      team.TeamId,
			PerformedBy: performedBy,
			Timestamp:   time.Now(),
		}
		s.kafkaProducer.ProduceTeamEvent(event)
	}

	return map[string]interface{}{
		"teamId":   team.TeamId,
		"teamName": team.TeamName,
		"managers": managers,
		"members":  members,
	}, nil
}

func (s *teamService) AddMember(teamID uint, memberID string, performedBy string) error {
	roster := &entities.Roster{
		TeamId:   teamID,
		UserId:   memberID,
		IsLeader: false,
	}

	err := s.teamRepo.CreateRoster(roster)
	if err != nil {
		return err
	}

	// Send Kafka event for member addition
	if s.kafkaProducer != nil {
		event := kafka.TeamEvent{
			EventType:    kafka.MemberAdded,
			TeamId:       teamID,
			PerformedBy:  performedBy,
			TargetUserId: memberID,
			Timestamp:    time.Now(),
		}
		s.kafkaProducer.ProduceTeamEvent(event)
	}

	return nil
}

func (s *teamService) DeleteMember(teamID uint, memberID string, performedBy string) error {
	err := s.teamRepo.DeleteRoster(teamID, memberID, false)
	if err != nil {
		return err
	}

	// Send Kafka event for member removal
	if s.kafkaProducer != nil {
		event := kafka.TeamEvent{
			EventType:    kafka.MemberRemoved,
			TeamId:       teamID,
			PerformedBy:  performedBy,
			TargetUserId: memberID,
			Timestamp:    time.Now(),
		}
		s.kafkaProducer.ProduceTeamEvent(event)
	}

	return nil
}

func (s *teamService) AddManager(teamID uint, managerID string, performedBy string) error {
	roster := &entities.Roster{
		TeamId:   teamID,
		UserId:   managerID,
		IsLeader: true,
	}

	err := s.teamRepo.CreateRoster(roster)
	if err != nil {
		return err
	}

	// Send Kafka event for manager addition
	if s.kafkaProducer != nil {
		event := kafka.TeamEvent{
			EventType:    kafka.ManagerAdded,
			TeamId:       teamID,
			PerformedBy:  performedBy,
			TargetUserId: managerID,
			Timestamp:    time.Now(),
		}
		s.kafkaProducer.ProduceTeamEvent(event)
	}

	return nil
}

func (s *teamService) DeleteManager(teamID uint, managerID string, performedBy string) error {
	err := s.teamRepo.DeleteRoster(teamID, managerID, true)
	if err != nil {
		return err
	}

	// Send Kafka event for manager removal
	if s.kafkaProducer != nil {
		event := kafka.TeamEvent{
			EventType:    kafka.ManagerRemoved,
			TeamId:       teamID,
			PerformedBy:  performedBy,
			TargetUserId: managerID,
			Timestamp:    time.Now(),
		}
		s.kafkaProducer.ProduceTeamEvent(event)
	}

	return nil
}

// Helper function to parse string to uint
func parseUint(s string) uint {
	var v uint
	fmt.Sscanf(s, "%d", &v)
	return v
}
