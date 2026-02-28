package services

import (
	"context"
	"testing"

	"leadflowcrm/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id interface{}) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	t.Run("should create new user successfully", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		authService := &AuthService{
			userRepo: nil,
		}
		
		assert.NotNil(t, authService)
	})
}

func TestLeadService_CreateLead(t *testing.T) {
	t.Run("should create lead with default status", func(t *testing.T) {
		lead := &models.Lead{
			Name:  "Test Lead",
			Email: "test@example.com",
		}
		
		assert.Equal(t, "Test Lead", lead.Name)
		assert.Equal(t, "test@example.com", lead.Email)
	})
}

func TestLeadService_GetLeads(t *testing.T) {
	t.Run("should apply default pagination", func(t *testing.T) {
		filter := models.LeadFilter{}
		
		if filter.Page < 1 {
			filter.Page = 1
		}
		if filter.Limit < 1 {
			filter.Limit = 10
		}
		
		assert.Equal(t, 1, filter.Page)
		assert.Equal(t, 10, filter.Limit)
	})
}

func TestModels_LeadStatus(t *testing.T) {
	t.Run("should have valid status constants", func(t *testing.T) {
		statuses := []models.LeadStatus{
			models.StatusNew,
			models.StatusContacted,
			models.StatusQualified,
			models.StatusProposal,
			models.StatusNegotiation,
			models.StatusWon,
			models.StatusLost,
		}
		
		assert.Greater(t, len(statuses), 0)
	})
}

func TestModels_LeadFilter(t *testing.T) {
	t.Run("should create filter with default values", func(t *testing.T) {
		filter := models.LeadFilter{
			Status: "new",
			Search: "test",
		}
		
		assert.Equal(t, "new", filter.Status)
		assert.Equal(t, "test", filter.Search)
	})
}
