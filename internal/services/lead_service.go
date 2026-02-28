package services

import (
	"context"
	"errors"
	"fmt"

	"leadflowcrm/internal/models"
	"leadflowcrm/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrLeadNotFound    = errors.New("lead not found")
	ErrUnauthorized    = errors.New("unauthorized")
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repositories.NewUserRepository(),
	}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	user := &models.User{
		Email:    req.Email,
		Password: string(hash),
		Name:     req.Name,
		Role:     role,
		Active:   true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidPassword
	}

	user.Password = ""
	return user, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	user.Password = ""
	return user, nil
}

func (s *AuthService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].Password = ""
	}
	return users, nil
}

type LeadService struct {
	leadRepo    *repositories.LeadRepository
	historyRepo *repositories.LeadHistoryRepository
	noteRepo    *repositories.LeadNoteRepository
	userRepo    *repositories.UserRepository
}

func NewLeadService() *LeadService {
	return &LeadService{
		leadRepo:    repositories.NewLeadRepository(),
		historyRepo: repositories.NewLeadHistoryRepository(),
		noteRepo:    repositories.NewLeadNoteRepository(),
		userRepo:    repositories.NewUserRepository(),
	}
}

func (s *LeadService) CreateLead(ctx context.Context, lead *models.Lead, userID primitive.ObjectID) (*models.Lead, error) {
	if err := s.leadRepo.Create(ctx, lead); err != nil {
		return nil, err
	}

	history := &models.LeadHistory{
		LeadID:      lead.ID,
		UserID:      userID,
		Action:      "created",
		Description: "Lead created",
	}
	s.historyRepo.Create(ctx, history)

	return lead, nil
}

func (s *LeadService) GetLead(ctx context.Context, id primitive.ObjectID) (*models.Lead, error) {
	lead, err := s.leadRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if lead == nil {
		return nil, ErrLeadNotFound
	}
	return lead, nil
}

func (s *LeadService) GetLeads(ctx context.Context, filter models.LeadFilter) ([]models.Lead, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	return s.leadRepo.Find(ctx, filter)
}

func (s *LeadService) UpdateLead(ctx context.Context, lead *models.Lead, userID primitive.ObjectID) error {
	oldLead, err := s.leadRepo.FindByID(ctx, lead.ID)
	if err != nil {
		return err
	}
	if oldLead == nil {
		return ErrLeadNotFound
	}

	if oldLead.Status != lead.Status {
		history := &models.LeadHistory{
			LeadID:      lead.ID,
			UserID:      userID,
			Action:      "status_change",
			Field:       "status",
			OldValue:    string(oldLead.Status),
			NewValue:    string(lead.Status),
			Description: fmt.Sprintf("Status changed from %s to %s", oldLead.Status, lead.Status),
		}
		s.historyRepo.Create(ctx, history)
	}

	if oldLead.AssignedTo != lead.AssignedTo {
		var oldUserName, newUserName string
		if !oldLead.AssignedTo.IsZero() {
			if u, _ := s.userRepo.FindByID(ctx, oldLead.AssignedTo); u != nil {
				oldUserName = u.Name
			}
		}
		if !lead.AssignedTo.IsZero() {
			if u, _ := s.userRepo.FindByID(ctx, lead.AssignedTo); u != nil {
				newUserName = u.Name
			}
		}
		history := &models.LeadHistory{
			LeadID:      lead.ID,
			UserID:      userID,
			Action:      "assignment_change",
			Field:       "assigned_to",
			OldValue:    oldUserName,
			NewValue:    newUserName,
			Description: fmt.Sprintf("Assignment changed from %s to %s", oldUserName, newUserName),
		}
		s.historyRepo.Create(ctx, history)
	}

	return s.leadRepo.Update(ctx, lead)
}

func (s *LeadService) DeleteLead(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	lead, err := s.leadRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if lead == nil {
		return ErrLeadNotFound
	}

	history := &models.LeadHistory{
		LeadID:      id,
		UserID:      userID,
		Action:      "deleted",
		Description: "Lead deleted",
	}
	s.historyRepo.Create(ctx, history)

	return s.leadRepo.SoftDelete(ctx, id)
}

func (s *LeadService) GetLeadHistory(ctx context.Context, leadID primitive.ObjectID) ([]models.LeadHistory, error) {
	return s.historyRepo.FindByLeadID(ctx, leadID)
}

func (s *LeadService) AddNote(ctx context.Context, note *models.LeadNote, userID primitive.ObjectID) (*models.LeadNote, error) {
	note.UserID = userID
	if err := s.noteRepo.Create(ctx, note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *LeadService) GetLeadNotes(ctx context.Context, leadID primitive.ObjectID) ([]models.LeadNote, error) {
	return s.noteRepo.FindByLeadID(ctx, leadID)
}

func (s *LeadService) DeleteNote(ctx context.Context, noteID primitive.ObjectID) error {
	return s.noteRepo.Delete(ctx, noteID)
}

func (s *LeadService) GetStats(ctx context.Context) (*models.LeadStats, error) {
	return s.leadRepo.GetStats(ctx)
}

func (s *LeadService) ImportCSV(ctx context.Context, leads []models.Lead, userID primitive.ObjectID) (int, error) {
	count := 0
	for i := range leads {
		if err := s.leadRepo.Create(ctx, &leads[i]); err != nil {
			continue
		}
		history := &models.LeadHistory{
			LeadID:      leads[i].ID,
			UserID:      userID,
			Action:      "imported",
			Description: "Lead imported from CSV",
		}
		s.historyRepo.Create(ctx, history)
		count++
	}
	return count, nil
}
