package apis_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hylarucoder/rocketbase/core"
	"github.com/hylarucoder/rocketbase/tests"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
)

func (suite *SettingsTestSuite) TestList() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodGet,
			Url:             "/api/settings",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as auth record",
			Method: http.MethodGet,
			Url:    "/api/settings",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin",
			Method: http.MethodGet,
			Url:    "/api/settings",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"meta":{`,
				`"logs":{`,
				`"smtp":{`,
				`"s3":{`,
				`"backups":{`,
				`"adminAuthToken":{`,
				`"adminPasswordResetToken":{`,
				`"adminFileToken":{`,
				`"recordAuthToken":{`,
				`"recordPasswordResetToken":{`,
				`"recordEmailChangeToken":{`,
				`"recordVerificationToken":{`,
				`"recordFileToken":{`,
				`"emailAuth":{`,
				`"googleAuth":{`,
				`"facebookAuth":{`,
				`"githubAuth":{`,
				`"gitlabAuth":{`,
				`"twitterAuth":{`,
				`"discordAuth":{`,
				`"microsoftAuth":{`,
				`"spotifyAuth":{`,
				`"kakaoAuth":{`,
				`"twitchAuth":{`,
				`"stravaAuth":{`,
				`"giteeAuth":{`,
				`"livechatAuth":{`,
				`"giteaAuth":{`,
				`"oidcAuth":{`,
				`"oidc2Auth":{`,
				`"oidc3Auth":{`,
				`"appleAuth":{`,
				`"instagramAuth":{`,
				`"vkAuth":{`,
				`"yandexAuth":{`,
				`"patreonAuth":{`,
				`"mailcowAuth":{`,
				`"secret":"******"`,
				`"clientSecret":"******"`,
			},
			ExpectedEvents: map[string]int{
				"OnSettingsListRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *SettingsTestSuite) TestSet() {
	t := suite.T()

	validData := `{"meta":{"appName":"update_test"}}`

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodPatch,
			Url:             "/api/settings",
			Body:            strings.NewReader(validData),
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as auth record",
			Method: http.MethodPatch,
			Url:    "/api/settings",
			Body:   strings.NewReader(validData),
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin submitting empty data",
			Method: http.MethodPatch,
			Url:    "/api/settings",
			Body:   strings.NewReader(``),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"meta":{`,
				`"logs":{`,
				`"smtp":{`,
				`"s3":{`,
				`"backups":{`,
				`"adminAuthToken":{`,
				`"adminPasswordResetToken":{`,
				`"adminFileToken":{`,
				`"recordAuthToken":{`,
				`"recordPasswordResetToken":{`,
				`"recordEmailChangeToken":{`,
				`"recordVerificationToken":{`,
				`"recordFileToken":{`,
				`"emailAuth":{`,
				`"googleAuth":{`,
				`"facebookAuth":{`,
				`"githubAuth":{`,
				`"gitlabAuth":{`,
				`"discordAuth":{`,
				`"microsoftAuth":{`,
				`"spotifyAuth":{`,
				`"kakaoAuth":{`,
				`"twitchAuth":{`,
				`"stravaAuth":{`,
				`"giteeAuth":{`,
				`"livechatAuth":{`,
				`"giteaAuth":{`,
				`"oidcAuth":{`,
				`"oidc2Auth":{`,
				`"oidc3Auth":{`,
				`"appleAuth":{`,
				`"instagramAuth":{`,
				`"vkAuth":{`,
				`"yandexAuth":{`,
				`"patreonAuth":{`,
				`"mailcowAuth":{`,
				`"secret":"******"`,
				`"clientSecret":"******"`,
				`"appName":"acme_test"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelBeforeUpdate":           1,
				"OnModelAfterUpdate":            1,
				"OnSettingsBeforeUpdateRequest": 1,
				"OnSettingsAfterUpdateRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin submitting invalid data",
			Method: http.MethodPatch,
			Url:    "/api/settings",
			Body:   strings.NewReader(`{"meta":{"appName":""}}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"meta":{"appName":{"code":"validation_required"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin submitting valid data",
			Method: http.MethodPatch,
			Url:    "/api/settings",
			Body:   strings.NewReader(validData),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"meta":{`,
				`"logs":{`,
				`"smtp":{`,
				`"s3":{`,
				`"backups":{`,
				`"adminAuthToken":{`,
				`"adminPasswordResetToken":{`,
				`"adminFileToken":{`,
				`"recordAuthToken":{`,
				`"recordPasswordResetToken":{`,
				`"recordEmailChangeToken":{`,
				`"recordVerificationToken":{`,
				`"recordFileToken":{`,
				`"emailAuth":{`,
				`"googleAuth":{`,
				`"facebookAuth":{`,
				`"githubAuth":{`,
				`"gitlabAuth":{`,
				`"twitterAuth":{`,
				`"discordAuth":{`,
				`"microsoftAuth":{`,
				`"spotifyAuth":{`,
				`"kakaoAuth":{`,
				`"twitchAuth":{`,
				`"stravaAuth":{`,
				`"giteeAuth":{`,
				`"livechatAuth":{`,
				`"giteaAuth":{`,
				`"oidcAuth":{`,
				`"oidc2Auth":{`,
				`"oidc3Auth":{`,
				`"appleAuth":{`,
				`"instagramAuth":{`,
				`"vkAuth":{`,
				`"yandexAuth":{`,
				`"patreonAuth":{`,
				`"mailcowAuth":{`,
				`"secret":"******"`,
				`"clientSecret":"******"`,
				`"appName":"update_test"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelBeforeUpdate":           1,
				"OnModelAfterUpdate":            1,
				"OnSettingsBeforeUpdateRequest": 1,
				"OnSettingsAfterUpdateRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "OnSettingsAfterUpdateRequest error response",
			Method: http.MethodPatch,
			Url:    "/api/settings",
			Body:   strings.NewReader(validData),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnSettingsAfterUpdateRequest().Add(func(e *core.SettingsUpdateEvent) error {
					return errors.New("error")
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnModelBeforeUpdate":           1,
				"OnModelAfterUpdate":            1,
				"OnSettingsBeforeUpdateRequest": 1,
				"OnSettingsAfterUpdateRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *SettingsTestSuite) TestTestS3() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodPost,
			Url:             "/api/settings/test/s3",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as auth record",
			Method: http.MethodPost,
			Url:    "/api/settings/test/s3",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (missing body + no s3)",
			Method: http.MethodPost,
			Url:    "/api/settings/test/s3",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"filesystem":{`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (invalid filesystem)",
			Method: http.MethodPost,
			Url:    "/api/settings/test/s3",
			Body:   strings.NewReader(`{"filesystem":"invalid"}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"filesystem":{`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (valid filesystem and no s3)",
			Method: http.MethodPost,
			Url:    "/api/settings/test/s3",
			Body:   strings.NewReader(`{"filesystem":"storage"}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *SettingsTestSuite) TestEmail() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:   "unauthorized",
			Method: http.MethodPost,
			Url:    "/api/settings/test/email",
			Body: strings.NewReader(`{
				"template": "verification",
				"email": "test@example.com"
			}`),
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as auth record",
			Method: http.MethodPost,
			Url:    "/api/settings/test/email",
			Body: strings.NewReader(`{
				"template": "verification",
				"email": "test@example.com"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (invalid body)",
			Method: http.MethodPost,
			Url:    "/api/settings/test/email",
			Body:   strings.NewReader(`{`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (empty json)",
			Method: http.MethodPost,
			Url:    "/api/settings/test/email",
			Body:   strings.NewReader(`{}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"email":{"code":"validation_required"`,
				`"template":{"code":"validation_required"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (verifiation template)",
			Method: http.MethodPost,
			Url:    "/api/settings/test/email",
			Body: strings.NewReader(`{
				"template": "verification",
				"email": "test@example.com"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				if app.TestMailer.TotalSend != 1 {
					t.Fatalf("[verification] Expected 1 sent email, got %d", app.TestMailer.TotalSend)
				}

				if len(app.TestMailer.LastMessage.To) != 1 {
					t.Fatalf("[verification] Expected 1 recipient, got %v", app.TestMailer.LastMessage.To)
				}

				if app.TestMailer.LastMessage.To[0].Address != "test@example.com" {
					t.Fatalf("[verification] Expected the email to be sent to %s, got %s", "test@example.com", app.TestMailer.LastMessage.To[0].Address)
				}

				if !strings.Contains(app.TestMailer.LastMessage.HTML, "Verify") {
					t.Fatalf("[verification] Expected to sent a verification email, got \n%v\n%v", app.TestMailer.LastMessage.Subject, app.TestMailer.LastMessage.HTML)
				}
			},
			ExpectedStatus:  204,
			ExpectedContent: []string{},
			ExpectedEvents: map[string]int{
				"OnMailerBeforeRecordVerificationSend": 1,
				"OnMailerAfterRecordVerificationSend":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (password reset template)",
			Method: http.MethodPost,
			Url:    "/api/settings/test/email",
			Body: strings.NewReader(`{
				"template": "password-reset",
				"email": "test@example.com"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				if app.TestMailer.TotalSend != 1 {
					t.Fatalf("[password-reset] Expected 1 sent email, got %d", app.TestMailer.TotalSend)
				}

				if len(app.TestMailer.LastMessage.To) != 1 {
					t.Fatalf("[password-reset] Expected 1 recipient, got %v", app.TestMailer.LastMessage.To)
				}

				if app.TestMailer.LastMessage.To[0].Address != "test@example.com" {
					t.Fatalf("[password-reset] Expected the email to be sent to %s, got %s", "test@example.com", app.TestMailer.LastMessage.To[0].Address)
				}

				if !strings.Contains(app.TestMailer.LastMessage.HTML, "Reset password") {
					t.Fatalf("[password-reset] Expected to sent a password-reset email, got \n%v\n%v", app.TestMailer.LastMessage.Subject, app.TestMailer.LastMessage.HTML)
				}
			},
			ExpectedStatus:  204,
			ExpectedContent: []string{},
			ExpectedEvents: map[string]int{
				"OnMailerBeforeRecordResetPasswordSend": 1,
				"OnMailerAfterRecordResetPasswordSend":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (email change)",
			Method: http.MethodPost,
			Url:    "/api/settings/test/email",
			Body: strings.NewReader(`{
				"template": "email-change",
				"email": "test@example.com"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				if app.TestMailer.TotalSend != 1 {
					t.Fatalf("[email-change] Expected 1 sent email, got %d", app.TestMailer.TotalSend)
				}

				if len(app.TestMailer.LastMessage.To) != 1 {
					t.Fatalf("[email-change] Expected 1 recipient, got %v", app.TestMailer.LastMessage.To)
				}

				if app.TestMailer.LastMessage.To[0].Address != "test@example.com" {
					t.Fatalf("[email-change] Expected the email to be sent to %s, got %s", "test@example.com", app.TestMailer.LastMessage.To[0].Address)
				}

				if !strings.Contains(app.TestMailer.LastMessage.HTML, "Confirm new email") {
					t.Fatalf("[email-change] Expected to sent a confirm new email email, got \n%v\n%v", app.TestMailer.LastMessage.Subject, app.TestMailer.LastMessage.HTML)
				}
			},
			ExpectedStatus:  204,
			ExpectedContent: []string{},
			ExpectedEvents: map[string]int{
				"OnMailerBeforeRecordChangeEmailSend": 1,
				"OnMailerAfterRecordChangeEmailSend":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *SettingsTestSuite) TestGenerateAppleClientSecret() {
	t := suite.T()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	encodedKey, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}

	privatePem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: encodedKey,
		},
	)

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodPost,
			Url:             "/api/settings/apple/generate-client-secret",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as auth record",
			Method: http.MethodPost,
			Url:    "/api/settings/apple/generate-client-secret",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (invalid body)",
			Method: http.MethodPost,
			Url:    "/api/settings/apple/generate-client-secret",
			Body:   strings.NewReader(`{`),
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (empty json)",
			Method: http.MethodPost,
			Url:    "/api/settings/apple/generate-client-secret",
			Body:   strings.NewReader(`{}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"clientId":{"code":"validation_required"`,
				`"teamId":{"code":"validation_required"`,
				`"keyId":{"code":"validation_required"`,
				`"privateKey":{"code":"validation_required"`,
				`"duration":{"code":"validation_required"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (invalid data)",
			Method: http.MethodPost,
			Url:    "/api/settings/apple/generate-client-secret",
			Body: strings.NewReader(`{
				"clientId": "",
				"teamId": "123456789",
				"keyId": "123456789",
				"privateKey": "invalid",
				"duration": -1
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"clientId":{"code":"validation_required"`,
				`"teamId":{"code":"validation_length_invalid"`,
				`"keyId":{"code":"validation_length_invalid"`,
				`"privateKey":{"code":"validation_match_invalid"`,
				`"duration":{"code":"validation_min_greater_equal_than_required"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
		{
			Name:   "authorized as admin (valid data)",
			Method: http.MethodPost,
			Url:    "/api/settings/apple/generate-client-secret",
			Body: strings.NewReader(fmt.Sprintf(`{
				"clientId": "123",
				"teamId": "1234567890",
				"keyId": "1234567891",
				"privateKey": %q,
				"duration": 1
			}`, privatePem)),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"secret":"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return suite.App
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

type SettingsTestSuite struct {
	suite.Suite
	App            *tests.TestApp
	AdminAuthToken string
	UserAuthToken  string
}

func (suite *SettingsTestSuite) SetupSuite() {
	app, _ := tests.NewTestApp()
	suite.AdminAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzAyMzYxMTQsImlkIjoiMjEwNzk3NzEyNzUyODc1OTI5NiIsInR5cGUiOiJhZG1pbiJ9.ikCEJR-iPIrZwpbsWjtslMdq75suCAEYfaRK7Oz-NZ0"
	suite.UserAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiIyMTA3OTc3Mzk3MDYzMTIyOTQ0IiwiZXhwIjoxNzMwOTEyMTQzLCJpZCI6Il9wYl91c2Vyc19hdXRoXyIsInR5cGUiOiJhdXRoUmVjb3JkIiwidmVyaWZpZWQiOnRydWV9.Us_731ziRkeeZvYvXiXsc6CKEwdKp4rSvsGbG5L1OUQ"
	suite.App = app
}

func (suite *SettingsTestSuite) TearDownSuite() {
	suite.App.Cleanup()
}

func TestRecordUpsertTestSuite(t *testing.T) {
	suite.Run(t, new(SettingsTestSuite))
}
