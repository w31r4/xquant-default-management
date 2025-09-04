package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"xquant-default-management/internal/api"
	"xquant-default-management/internal/core"
	"xquant-default-management/internal/handler"
	"xquant-default-management/internal/middleware"
	"xquant-default-management/internal/repository"
	"xquant-default-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Re-using the suite from the other integration test to manage the DB connection
type ApiEndpointIntegrationSuite struct {
	ServiceRepoIntegrationSuite // Embed the previous suite
	server                      *httptest.Server
}

func (s *ApiEndpointIntegrationSuite) SetupSuite() {
	// Call the embedded suite's setup
	s.ServiceRepoIntegrationSuite.SetupSuite()

	// Setup a full router like in main.go
	router := gin.Default()

	// Dependency Injection from main.go (simplified)
	userRepo := repository.NewUserRepository(s.db)
	appRepo := repository.NewApplicationRepository(s.db)
	customerRepo := repository.NewCustomerRepository(s.db)
	statsRepo := repository.NewStatisticsRepository(s.db)

	userService := service.NewUserService(userRepo, s.cfg)
	appService := service.NewApplicationService(s.db, appRepo, customerRepo)
	queryService := service.NewQueryService(appRepo)
	statsService := service.NewStatisticsService(statsRepo)

	userHandler := handler.NewUserHandler(userService)
	appHandler := handler.NewApplicationHandler(appService)
	_ = handler.NewQueryHandler(queryService)
	_ = handler.NewStatisticsHandler(statsService)

	// Setup routes
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/register", userHandler.Register)
		apiV1.POST("/login", userHandler.Login)

		protected := apiV1.Group("/")
		protected.Use(middleware.AuthMiddleware(s.cfg))
		{
			applications := protected.Group("/applications")
			{
				applications.POST("", middleware.RBACMiddleware("Applicant"), appHandler.CreateApplication)
				applications.GET("/pending", middleware.RBACMiddleware("Approver"), appHandler.GetPendingApplications)
				review := applications.Group("/review")
				review.Use(middleware.RBACMiddleware("Approver"))
				{
					review.POST("/approve", appHandler.ApproveApplication)
				}
			}
		}
	}

	s.server = httptest.NewServer(router)
}

func (s *ApiEndpointIntegrationSuite) TearDownSuite() {
	s.server.Close()
	s.ServiceRepoIntegrationSuite.TearDownSuite()
}

func TestApiEndpointIntegration(t *testing.T) {
	suite.Run(t, new(ApiEndpointIntegrationSuite))
}

func (s *ApiEndpointIntegrationSuite) TestUserRegisterAndLoginEndpoints() {
	// 1. Test Registration Endpoint
	registerReqBody := api.RegisterRequest{
		Username: "endpoint_user",
		Password: "password123",
		Role:     "Applicant",
	}
	registerJson, _ := json.Marshal(registerReqBody)

	resp, err := http.Post(s.server.URL+"/api/v1/register", "application/json", bytes.NewBuffer(registerJson))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	var registerRes api.UserResponse
	json.NewDecoder(resp.Body).Decode(&registerRes)
	assert.Equal(s.T(), registerReqBody.Username, registerRes.Username)

	// 2. Test Login Endpoint
	loginReqBody := api.LoginRequest{
		Username: "endpoint_user",
		Password: "password123",
	}
	loginJson, _ := json.Marshal(loginReqBody)

	resp, err = http.Post(s.server.URL+"/api/v1/login", "application/json", bytes.NewBuffer(loginJson))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)
	var loginRes api.LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginRes)
	assert.NotEmpty(s.T(), loginRes.Token)
}

func (s *ApiEndpointIntegrationSuite) TestFullApplicationE2EWorkflow() {
	// Step 1: Register Applicant and Approver
	applicantUsername := "e2e_applicant"
	approverUsername := "e2e_approver"
	password := "secure_password"

	s.registerUser(applicantUsername, password, "Applicant")
	s.registerUser(approverUsername, password, "Approver")

	// Step 2: Applicant logs in
	applicantToken := s.loginUser(applicantUsername, password)

	// Step 3: Applicant creates a default application
	// (Assuming a customer already exists, or create one here)
	customer := core.Customer{Name: "E2E Test Customer", Industry: "Tech", Region: "NA"}
	s.db.Create(&customer)

	createAppReq := api.CreateApplicationRequest{
		CustomerName: customer.Name,
		Severity:     "High",
		Reason:       "E2E Test Reason",
	}
	createAppJson, _ := json.Marshal(createAppReq)
	req, _ := http.NewRequest(http.MethodPost, s.server.URL+"/api/v1/applications", bytes.NewBuffer(createAppJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+applicantToken)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	var createdApp api.ApplicationResponse
	json.NewDecoder(resp.Body).Decode(&createdApp)
	resp.Body.Close()

	// Step 4: Approver logs in
	approverToken := s.loginUser(approverUsername, password)

	// Step 5: Approver gets pending applications
	req, _ = http.NewRequest(http.MethodGet, s.server.URL+"/api/v1/applications/pending", nil)
	req.Header.Set("Authorization", "Bearer "+approverToken)
	resp, err = http.DefaultClient.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	var pendingApps []api.ApplicationResponse
	json.NewDecoder(resp.Body).Decode(&pendingApps)
	resp.Body.Close()
	s.Assert().NotEmpty(pendingApps, "Pending applications list should not be empty")
	s.Assert().Equal(createdApp.ID, pendingApps[0].ID)

	// Step 6: Approver approves the application
	approveReq := api.ApproveRequest{ApplicationID: createdApp.ID}
	approveJson, _ := json.Marshal(approveReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.URL+"/api/v1/applications/review/approve", bytes.NewBuffer(approveJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+approverToken)
	resp, err = http.DefaultClient.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Step 7: Verify application status
	var finalApp core.DefaultApplication
	err = s.db.First(&finalApp, "id = ?", createdApp.ID).Error
	s.Require().NoError(err)
	s.Assert().Equal("Approved", finalApp.Status)
}

// Helper functions to reduce code duplication
func (s *ApiEndpointIntegrationSuite) registerUser(username, password, role string) {
	reqBody := api.RegisterRequest{Username: username, Password: password, Role: role}
	jsonBody, _ := json.Marshal(reqBody)
	resp, err := http.Post(s.server.URL+"/api/v1/register", "application/json", bytes.NewBuffer(jsonBody))
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	resp.Body.Close()
}

func (s *ApiEndpointIntegrationSuite) loginUser(username, password string) string {
	reqBody := api.LoginRequest{Username: username, Password: password}
	jsonBody, _ := json.Marshal(reqBody)
	resp, err := http.Post(s.server.URL+"/api/v1/login", "application/json", bytes.NewBuffer(jsonBody))
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	var loginRes api.LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginRes)
	resp.Body.Close()
	s.Require().NotEmpty(loginRes.Token)
	return loginRes.Token
}
