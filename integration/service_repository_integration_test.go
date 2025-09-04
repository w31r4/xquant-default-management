package integration

import (
	"log"
	"os"
	"testing"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/core"
	"xquant-default-management/internal/database"
	"xquant-default-management/internal/repository"
	"xquant-default-management/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ServiceRepoIntegrationSuite struct {
	suite.Suite
	db          *gorm.DB
	cfg         config.Config
	userService service.UserService
	// Add other services and repos as needed
}

func (s *ServiceRepoIntegrationSuite) SetupSuite() {
	// Load test config
	cfg, err := config.LoadConfig("../configs")
	if err != nil {
		log.Fatalf("Failed to load test config: %v", err)
	}
	s.cfg = cfg

	// Connect to the test database
	database.Connect(cfg)
	s.db = database.DB

	// Auto-migrate the schema
	err = s.db.AutoMigrate(&core.User{}, &core.Customer{}, &core.DefaultApplication{})
	s.Require().NoError(err)

	// Initialize real repositories and services
	userRepo := repository.NewUserRepository(s.db)
	s.userService = service.NewUserService(userRepo, s.cfg)
}

func (s *ServiceRepoIntegrationSuite) TearDownSuite() {
	// Close the database connection
	sqlDB, _ := s.db.DB()
	sqlDB.Close()
}

func (s *ServiceRepoIntegrationSuite) BeforeTest(suiteName, testName string) {
	// Clean up the database before each test
	s.db.Exec("DELETE FROM default_applications")
	s.db.Exec("DELETE FROM customers")
	s.db.Exec("DELETE FROM users")
}

func TestServiceRepoIntegration(t *testing.T) {
	// This check prevents the integration test from running on CI/CD environments
	// where the database might not be available in the same way.
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration tests in CI environment")
	}
	suite.Run(t, new(ServiceRepoIntegrationSuite))
}

func (s *ServiceRepoIntegrationSuite) TestUserRegistrationAndLogin() {
	// 1. Register a new user
	username := "integration_user"
	password := "strong_password_123"
	role := "Applicant"
	user, err := s.userService.Register(username, password, role)

	s.T().Run("Register User", func(t *testing.T) {
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, username, user.Username)

		// Verify the user is actually in the database
		var dbUser core.User
		err := s.db.Where("username = ?", username).First(&dbUser).Error
		assert.NoError(t, err)
		assert.Equal(t, user.ID, dbUser.ID)
	})

	// 2. Login with the new user
	s.T().Run("Login User", func(t *testing.T) {
		token, err := s.userService.Login(username, password)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Try logging in with a wrong password
		_, err = s.userService.Login(username, "wrong_password")
		assert.Error(t, err)
	})
}
