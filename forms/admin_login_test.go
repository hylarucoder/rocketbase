package forms_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hylarucoder/rocketbase/forms"
	"github.com/hylarucoder/rocketbase/models"
	"github.com/hylarucoder/rocketbase/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func (s *AdminLoginTestSuite) TestAdminLoginValidateAndSubmit() {
	app := s.App
	form := forms.NewAdminLogin(app)

	scenarios := []struct {
		email       string
		password    string
		expectError bool
	}{
		{"", "", true},
		{"", "1234567890", true},
		{"test@example.com", "", true},
		{"test", "test", true},
		{"missing@example.com", "1234567890", true},
		{"test@example.com", "123456789", true},
		{"test@example.com", "1234567890", false},
	}

	for _, scenario := range scenarios {
		form.Identity = scenario.email
		form.Password = scenario.password

		admin, err := form.Submit()

		hasErr := err != nil
		assert.Equal(s.T(), scenario.expectError, hasErr)
		assert.Equal(s.T(), scenario.expectError, admin == nil)
		assert.Equal(s.T(), scenario.email, admin.Email)
	}
}

func (s *AdminLoginTestSuite) TestAdminLoginInterceptors() {
	testApp := s.App
	form := forms.NewAdminLogin(testApp)
	form.Identity = "test@example.com"
	form.Password = "123456"
	var interceptorAdmin *models.Admin
	testErr := errors.New("test_error")

	interceptor1Called := false
	interceptor1 := func(next forms.InterceptorNextFunc[*models.Admin]) forms.InterceptorNextFunc[*models.Admin] {
		return func(admin *models.Admin) error {
			interceptor1Called = true
			return next(admin)
		}
	}

	interceptor2Called := false
	interceptor2 := func(next forms.InterceptorNextFunc[*models.Admin]) forms.InterceptorNextFunc[*models.Admin] {
		return func(admin *models.Admin) error {
			interceptorAdmin = admin
			interceptor2Called = true
			return testErr
		}
	}

	_, submitErr := form.Submit(interceptor1, interceptor2)
	assert.Equal(s.T(), testErr, submitErr)

	assert.True(s.T(), interceptor1Called)
	assert.True(s.T(), interceptor2Called)

	assert.Equal(s.T(), form.Identity, interceptorAdmin.Email)
}

type AdminLoginTestSuite struct {
	suite.Suite
	App *tests.TestApp
	Var int
}

func (s *AdminLoginTestSuite) SetupTest() {
	app, _ := tests.NewTestApp()
	s.Var = 5
	s.App = app
}

func (s *AdminLoginTestSuite) TearDownTest() {
	s.App.Cleanup()
}

func (s *AdminLoginTestSuite) SetupSuite() {
	fmt.Println("setup suite")
}

func (s *AdminLoginTestSuite) TearDownSuite() {
	fmt.Println("teardown suite")
}

func TestAdminLoginTestSuite(t *testing.T) {
	suite.Run(t, new(AdminLoginTestSuite))
}
