package forms_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hylarucoder/rocketbase/forms"
	"github.com/hylarucoder/rocketbase/models"
	"github.com/hylarucoder/rocketbase/tests"
	"github.com/hylarucoder/rocketbase/tools/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func (suite *AdminPasswordResetConfirmTestSuite) TestAdminPasswordResetConfirmValidateAndSubmit() {
	app := suite.App

	form := forms.NewAdminPasswordResetConfirm(app)

	scenarios := []struct {
		token           string
		password        string
		passwordConfirm string
		expectError     bool
	}{
		{"", "", "", true},
		{"", "123", "", true},
		{"", "", "123", true},
		{"test", "", "", true},
		{"test", "123", "", true},
		{"test", "123", "123", true},
		{
			// expired
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhZG1pbiIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTY0MDk5MTY2MX0.GLwCOsgWTTEKXTK-AyGW838de1OeZGIjfHH0FoRLqZg",
			"1234567890",
			"1234567890",
			true,
		},
		{
			// valid with mismatched passwords
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhZG1pbiIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MjIwODk4MTYwMH0.kwFEler6KSMKJNstuaSDvE1QnNdCta5qSnjaIQ0hhhc",
			"1234567890",
			"1234567891",
			true,
		},
		{
			// valid with matching passwords
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhZG1pbiIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MjIwODk4MTYwMH0.kwFEler6KSMKJNstuaSDvE1QnNdCta5qSnjaIQ0hhhc",
			"1234567891",
			"1234567891",
			false,
		},
	}

	for _, s := range scenarios {
		form.Token = s.token
		form.Password = s.password
		form.PasswordConfirm = s.passwordConfirm

		interceptorCalls := 0
		interceptor := func(next forms.InterceptorNextFunc[*models.Admin]) forms.InterceptorNextFunc[*models.Admin] {
			return func(m *models.Admin) error {
				interceptorCalls++
				return next(m)
			}
		}

		admin, err := form.Submit(interceptor)

		// check interceptor calls
		expectInterceptorCalls := 1
		if s.expectError {
			expectInterceptorCalls = 0
		}
		assert.Equal(suite.T(), expectInterceptorCalls, interceptorCalls)

		assert.Equal(suite.T(), s.expectError, err != nil)

		if s.expectError {
			continue
		}

		claims, _ := security.ParseUnverifiedJWT(s.token)
		tokenAdminId := claims["id"]

		assert.Equal(suite.T(), tokenAdminId, admin.Id)

		assert.True(suite.T(), admin.ValidatePassword(form.Password))
	}
}

func (s *AdminPasswordResetConfirmTestSuite) TestAdminPasswordResetConfirmInterceptors() {
	testApp := s.App

	admin, err := testApp.Dao().FindAdminByEmail("test@example.com")
	if err != nil {
		s.T().Fatal(err)
	}

	form := forms.NewAdminPasswordResetConfirm(testApp)
	form.Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhZG1pbiIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MjIwODk4MTYwMH0.kwFEler6KSMKJNstuaSDvE1QnNdCta5qSnjaIQ0hhhc"
	form.Password = "1234567891"
	form.PasswordConfirm = "1234567891"
	interceptorTokenKey := admin.TokenKey
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
			interceptorTokenKey = admin.TokenKey
			interceptor2Called = true
			return testErr
		}
	}

	_, submitErr := form.Submit(interceptor1, interceptor2)
	assert.Equal(s.T(), testErr, submitErr)

	assert.True(s.T(), interceptor1Called)
	assert.True(s.T(), interceptor2Called)
	assert.NotEqual(s.T(), admin.TokenKey, interceptorTokenKey)
}

type AdminPasswordResetConfirmTestSuite struct {
	suite.Suite
	App *tests.TestApp
	Var int
}

func (s *AdminPasswordResetConfirmTestSuite) SetupTest() {
	app, _ := tests.NewTestApp()
	s.Var = 5
	s.App = app
}

func (s *AdminPasswordResetConfirmTestSuite) TearDownTest() {
	s.App.Cleanup()
}

func (s *AdminPasswordResetConfirmTestSuite) SetupSuite() {
	fmt.Println("setup suite")
}

func (s *AdminPasswordResetConfirmTestSuite) TearDownSuite() {
	fmt.Println("teardown suite")
}

func TestAdminPasswordResetConfirmTestSuite(t *testing.T) {
	suite.Run(t, new(AdminPasswordResetConfirmTestSuite))
}
