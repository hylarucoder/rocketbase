package apis_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hylarucoder/rocketbase/core"
	"github.com/hylarucoder/rocketbase/daos"
	"github.com/hylarucoder/rocketbase/models"
	"github.com/hylarucoder/rocketbase/tests"
	"github.com/hylarucoder/rocketbase/tools/types"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/suite"
)

func (suite *AdminTestSuite) TestAdminAuthWithPassword() {
	app := suite.App

	scenarios := []tests.ApiScenario{
		{
			Name:            "empty data",
			Method:          http.MethodPost,
			Url:             "/api/admins/auth-with-password",
			Body:            strings.NewReader(``),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{"identity":{"code":"validation_required","message":"Cannot be blank."},"password":{"code":"validation_required","message":"Cannot be blank."}}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:            "invalid data",
			Method:          http.MethodPost,
			Url:             "/api/admins/auth-with-password",
			Body:            strings.NewReader(`{`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:            "wrong email",
			Method:          http.MethodPost,
			Url:             "/api/admins/auth-with-password",
			Body:            strings.NewReader(`{"identity":"missing@example.com","password":"1234567890"}`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnAdminBeforeAuthWithPasswordRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:            "wrong password",
			Method:          http.MethodPost,
			Url:             "/api/admins/auth-with-password",
			Body:            strings.NewReader(`{"identity":"test@example.com","password":"invalid"}`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnAdminBeforeAuthWithPasswordRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:           "valid email/password (guest)",
			Method:         http.MethodPost,
			Url:            "/api/admins/auth-with-password",
			Body:           strings.NewReader(`{"identity":"test@example.com","password":"1234567890"}`),
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"admin":{"id":"2107977127528759297"`,
				`"token":`,
			},
			ExpectedEvents: map[string]int{
				"OnAdminBeforeAuthWithPasswordRequest": 1,
				"OnAdminAfterAuthWithPasswordRequest":  1,
				"OnAdminAuthRequest":                   1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "valid email/password (already authorized)",
			Method: http.MethodPost,
			Url:    "/api/admins/auth-with-password",
			Body:   strings.NewReader(`{"identity":"test@example.com","password":"1234567890"}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"admin":{"id":"2107977127528759297"`,
				`"token":`,
			},
			ExpectedEvents: map[string]int{
				"OnAdminBeforeAuthWithPasswordRequest": 1,
				"OnAdminAfterAuthWithPasswordRequest":  1,
				"OnAdminAuthRequest":                   1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "OnAdminAfterAuthWithPasswordRequest error response",
			Method: http.MethodPost,
			Url:    "/api/admins/auth-with-password",
			Body:   strings.NewReader(`{"identity":"test@example.com","password":"1234567890"}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnAdminAfterAuthWithPasswordRequest().Add(func(e *core.AdminAuthWithPasswordEvent) error {
					return errors.New("error")
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnAdminBeforeAuthWithPasswordRequest": 1,
				"OnAdminAfterAuthWithPasswordRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(suite.T())
	}
}

func (suite *AdminTestSuite) TestAdminRequestPasswordReset() {
	app := suite.App

	scenarios := []tests.ApiScenario{
		{
			Name:            "empty data",
			Method:          http.MethodPost,
			Url:             "/api/admins/request-password-reset",
			Body:            strings.NewReader(``),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{"email":{"code":"validation_required","message":"Cannot be blank."}}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:            "invalid data",
			Method:          http.MethodPost,
			Url:             "/api/admins/request-password-reset",
			Body:            strings.NewReader(`{"email`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:           "missing admin",
			Method:         http.MethodPost,
			Url:            "/api/admins/request-password-reset",
			Body:           strings.NewReader(`{"email":"missing@example.com"}`),
			Delay:          100 * time.Millisecond,
			ExpectedStatus: 204,
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:           "existing admin",
			Method:         http.MethodPost,
			Url:            "/api/admins/request-password-reset",
			Body:           strings.NewReader(`{"email":"test@example.com"}`),
			Delay:          100 * time.Millisecond,
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnModelBeforeUpdate":                      1,
				"OnModelAfterUpdate":                       1,
				"OnMailerBeforeAdminResetPasswordSend":     1,
				"OnMailerAfterAdminResetPasswordSend":      1,
				"OnAdminBeforeRequestPasswordResetRequest": 1,
				"OnAdminAfterRequestPasswordResetRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:           "existing admin (after already sent)",
			Method:         http.MethodPost,
			Url:            "/api/admins/request-password-reset",
			Body:           strings.NewReader(`{"email":"test@example.com"}`),
			Delay:          100 * time.Millisecond,
			ExpectedStatus: 204,
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				// simulate recent password request
				admin, err := app.Dao().FindAdminByEmail("test@example.com")
				if err != nil {
					t.Fatal(err)
				}
				admin.LastResetSentAt = types.NowDateTime()
				dao := daos.New(app.Dao().DB()) // new dao to ignore hooks
				if err := dao.Save(admin); err != nil {
					t.Fatal(err)
				}
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(suite.T())
	}
}

func (suite *AdminTestSuite) TestAdminConfirmPasswordReset() {
	app := suite.App

	scenarios := []tests.ApiScenario{
		{
			Name:            "empty data",
			Method:          http.MethodPost,
			Url:             "/api/admins/confirm-password-reset",
			Body:            strings.NewReader(``),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{"password":{"code":"validation_required","message":"Cannot be blank."},"passwordConfirm":{"code":"validation_required","message":"Cannot be blank."},"token":{"code":"validation_required","message":"Cannot be blank."}}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:            "invalid data",
			Method:          http.MethodPost,
			Url:             "/api/admins/confirm-password-reset",
			Body:            strings.NewReader(`{"password`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "expired token",
			Method: http.MethodPost,
			Url:    "/api/admins/confirm-password-reset",
			Body: strings.NewReader(`{
				"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhZG1pbiIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTY0MDk5MTY2MX0.GLwCOsgWTTEKXTK-AyGW838de1OeZGIjfHH0FoRLqZg",
				"password":"1234567890",
				"passwordConfirm":"1234567890"
			}`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{"token":{"code":"validation_invalid_token","message":"Invalid or expired token."}}}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "valid token + invalid password",
			Method: http.MethodPost,
			Url:    "/api/admins/confirm-password-reset",
			Body: strings.NewReader(`{
				"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhZG1pbiIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MjIwODk4MTYwMH0.kwFEler6KSMKJNstuaSDvE1QnNdCta5qSnjaIQ0hhhc",
				"password":"123456",
				"passwordConfirm":"123456"
			}`),
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{"password":{"code":"validation_length_out_of_range"`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "valid token + valid password",
			Method: http.MethodPost,
			Url:    "/api/admins/confirm-password-reset",
			// TODO: use new token?
			Body: strings.NewReader(`{
				"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhZG1pbiIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MjIwODk4MTYwMH0.kwFEler6KSMKJNstuaSDvE1QnNdCta5qSnjaIQ0hhhc",
				"password":"1234567891",
				"passwordConfirm":"1234567891"
			}`),
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnModelBeforeUpdate":                      1,
				"OnModelAfterUpdate":                       1,
				"OnAdminBeforeConfirmPasswordResetRequest": 1,
				"OnAdminAfterConfirmPasswordResetRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "OnAdminAfterConfirmPasswordResetRequest error response",
			Method: http.MethodPost,
			Url:    "/api/admins/confirm-password-reset",
			Body: strings.NewReader(`{
				"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhZG1pbiIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MjIwODk4MTYwMH0.kwFEler6KSMKJNstuaSDvE1QnNdCta5qSnjaIQ0hhhc",
				"password":"1234567891",
				"passwordConfirm":"1234567891"
			}`),
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnAdminAfterConfirmPasswordResetRequest().Add(func(e *core.AdminConfirmPasswordResetEvent) error {
					return errors.New("error")
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnModelBeforeUpdate":                      1,
				"OnModelAfterUpdate":                       1,
				"OnAdminBeforeConfirmPasswordResetRequest": 1,
				"OnAdminAfterConfirmPasswordResetRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(suite.T())
	}
}

func (suite *AdminTestSuite) TestAdminRefresh() {
	app := suite.App

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodPost,
			Url:             "/api/admins/auth-refresh",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as user",
			Method: http.MethodPost,
			Url:    "/api/admins/auth-refresh",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin (expired token)",
			Method: http.MethodPost,
			Url:    "/api/admins/auth-refresh",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin (valid token)",
			Method: http.MethodPost,
			Url:    "/api/admins/auth-refresh",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"admin":{"id":"2107977127528759297"`,
				`"token":`,
			},
			ExpectedEvents: map[string]int{
				"OnAdminAuthRequest":              1,
				"OnAdminBeforeAuthRefreshRequest": 1,
				"OnAdminAfterAuthRefreshRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "OnAdminAfterAuthRefreshRequest error response",
			Method: http.MethodPost,
			Url:    "/api/admins/auth-refresh",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnAdminAfterAuthRefreshRequest().Add(func(e *core.AdminAuthRefreshEvent) error {
					return errors.New("error")
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnAdminBeforeAuthRefreshRequest": 1,
				"OnAdminAfterAuthRefreshRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(suite.T())
	}
}

func (suite *AdminTestSuite) TestAdminsList() {
	app := suite.App

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodGet,
			Url:             "/api/admins",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as user",
			Method: http.MethodGet,
			Url:    "/api/admins",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin",
			Method: http.MethodGet,
			Url:    "/api/admins",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalItems":3`,
				`"items":[{`,
				`"id":"2107977127528759297"`,
				`"id":"2107977127528759298"`,
				`"id":"9q2trqumvlyr3bd"`,
			},
			ExpectedEvents: map[string]int{
				"OnAdminsListRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + paging and sorting",
			Method: http.MethodGet,
			Url:    "/api/admins?page=2&perPage=1&sort=-created",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":2`,
				`"perPage":1`,
				`"totalItems":3`,
				`"items":[{`,
				`"id":"2107977127528759298"`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnAdminsListRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + invalid filter",
			Method: http.MethodGet,
			Url:    "/api/admins?filter=invalidfield~'test2'",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + valid filter",
			Method: http.MethodGet,
			Url:    "/api/admins?filter=email~'test3'",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalItems":1`,
				`"items":[{`,
				`"id":"9q2trqumvlyr3bd"`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnAdminsListRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(suite.T())
	}
}

func (suite *AdminTestSuite) TestAdminView() {
	app := suite.App

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodGet,
			Url:             "/api/admins/2107977127528759298",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as user",
			Method: http.MethodGet,
			Url:    "/api/admins/2107977127528759298",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + nonexisting admin id",
			Method: http.MethodGet,
			Url:    "/api/admins/nonexisting",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + existing admin id",
			Method: http.MethodGet,
			Url:    "/api/admins/2107977127528759298",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"2107977127528759298"`,
			},
			NotExpectedContent: []string{
				`"tokenKey"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnAdminViewRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(suite.T())
	}
}

func (suite *AdminTestSuite) TestAdminDelete() {
	app := suite.App

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodDelete,
			Url:             "/api/admins/2107977127528759298",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as user",
			Method: http.MethodDelete,
			Url:    "/api/admins/2107977127528759298",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + missing admin id",
			Method: http.MethodDelete,
			Url:    "/api/admins/missing",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + existing admin id",
			Method: http.MethodDelete,
			Url:    "/api/admins/2107977127528759298",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnModelBeforeDelete":        1,
				"OnModelAfterDelete":         1,
				"OnAdminBeforeDeleteRequest": 1,
				"OnAdminAfterDeleteRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin - try to delete the only remaining admin",
			Method: http.MethodDelete,
			Url:    "/api/admins/2107977127528759297",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				// delete all admins except the authorized one
				adminModel := &models.Admin{}
				_, err := app.Dao().DB().Delete(adminModel.TableName(), dbx.Not(dbx.HashExp{
					"id": "2107977127528759297",
				})).Execute()
				if err != nil {
					t.Fatal(err)
				}
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnAdminBeforeDeleteRequest": 1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "OnAdminAfterDeleteRequest error response",
			Method: http.MethodDelete,
			Url:    "/api/admins/2107977127528759298",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnAdminAfterDeleteRequest().Add(func(e *core.AdminDeleteEvent) error {
					return errors.New("error")
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnModelBeforeDelete":        1,
				"OnModelAfterDelete":         1,
				"OnAdminBeforeDeleteRequest": 1,
				"OnAdminAfterDeleteRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(suite.T())
	}
}

func (suite *AdminTestSuite) TestAdminCreate() {
	app := suite.App

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized (while having at least 1 existing admin)",
			Method:          http.MethodPost,
			Url:             "/api/admins",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as user",
			Method: http.MethodPost,
			Url:    "/api/admins",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + empty data",
			Method: http.MethodPost,
			Url:    "/api/admins",
			Body:   strings.NewReader(``),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{"email":{"code":"validation_required","message":"Cannot be blank."},"password":{"code":"validation_required","message":"Cannot be blank."}}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + invalid data format",
			Method: http.MethodPost,
			Url:    "/api/admins",
			Body:   strings.NewReader(`{`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + invalid data",
			Method: http.MethodPost,
			Url:    "/api/admins",
			Body: strings.NewReader(`{
				"email":"test@example.com",
				"password":"1234",
				"passwordConfirm":"4321",
				"avatar":99
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"avatar":{"code":"validation_max_less_equal_than_required"`,
				`"email":{"code":"validation_admin_email_exists"`,
				`"password":{"code":"validation_length_out_of_range"`,
				`"passwordConfirm":{"code":"validation_values_mismatch"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + valid data",
			Method: http.MethodPost,
			Url:    "/api/admins",
			Body: strings.NewReader(`{
				"email":"testnew@example.com",
				"password":"1234567890",
				"passwordConfirm":"1234567890",
				"avatar":3
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":`,
				`"email":"testnew@example.com"`,
				`"avatar":3`,
			},
			NotExpectedContent: []string{
				`"password"`,
				`"passwordConfirm"`,
				`"tokenKey"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelBeforeCreate":        1,
				"OnModelAfterCreate":         1,
				"OnAdminBeforeCreateRequest": 1,
				"OnAdminAfterCreateRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "OnAdminAfterCreateRequest error response",
			Method: http.MethodPost,
			Url:    "/api/admins",
			Body: strings.NewReader(`{
				"email":"testnew@example.com",
				"password":"1234567890",
				"passwordConfirm":"1234567890",
				"avatar":3
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnAdminAfterCreateRequest().Add(func(e *core.AdminCreateEvent) error {
					return errors.New("error")
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnModelBeforeCreate":        1,
				"OnModelAfterCreate":         1,
				"OnAdminBeforeCreateRequest": 1,
				"OnAdminAfterCreateRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "unauthorized (while having 0 existing admins)",
			Method: http.MethodPost,
			Url:    "/api/admins",
			Body:   strings.NewReader(`{"email":"testnew@example.com","password":"1234567890","passwordConfirm":"1234567890","avatar":3}`),
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				// delete all admins
				_, err := app.Dao().DB().NewQuery("DELETE FROM {{_admins}}").Execute()
				if err != nil {
					t.Fatal(err)
				}
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":`,
				`"email":"testnew@example.com"`,
				`"avatar":3`,
			},
			ExpectedEvents: map[string]int{
				"OnModelBeforeCreate":        1,
				"OnModelAfterCreate":         1,
				"OnAdminBeforeCreateRequest": 1,
				"OnAdminAfterCreateRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(suite.T())
	}
}

func (suite *AdminTestSuite) TestAdminUpdate() {
	app := suite.App

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodPatch,
			Url:             "/api/admins/2107977127528759298",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as user",
			Method: http.MethodPatch,
			Url:    "/api/admins/2107977127528759298",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + missing admin",
			Method: http.MethodPatch,
			Url:    "/api/admins/missing",
			Body:   strings.NewReader(``),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + empty data",
			Method: http.MethodPatch,
			Url:    "/api/admins/2107977127528759298",
			Body:   strings.NewReader(``),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"2107977127528759298"`,
				`"email":"test2@example.com"`,
				`"avatar":2`,
			},
			ExpectedEvents: map[string]int{
				"OnModelBeforeUpdate":        1,
				"OnModelAfterUpdate":         1,
				"OnAdminBeforeUpdateRequest": 1,
				"OnAdminAfterUpdateRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + invalid formatted data",
			Method: http.MethodPatch,
			Url:    "/api/admins/2107977127528759298",
			Body:   strings.NewReader(`{`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + invalid data",
			Method: http.MethodPatch,
			Url:    "/api/admins/2107977127528759298",
			Body: strings.NewReader(`{
				"email":"test@example.com",
				"password":"1234",
				"passwordConfirm":"4321",
				"avatar":99
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{`,
				`"avatar":{"code":"validation_max_less_equal_than_required"`,
				`"email":{"code":"validation_admin_email_exists"`,
				`"password":{"code":"validation_length_out_of_range"`,
				`"passwordConfirm":{"code":"validation_values_mismatch"`,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "authorized as admin + valid data",
			Method: http.MethodPatch,
			Url:    "/api/admins/2107977127528759298",
			Body: strings.NewReader(`{
				"email":"testnew@example.com",
				"password":"1234567891",
				"passwordConfirm":"1234567891",
				"avatar":5
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"2107977127528759298"`,
				`"email":"testnew@example.com"`,
				`"avatar":5`,
			},
			NotExpectedContent: []string{
				`"password"`,
				`"passwordConfirm"`,
				`"tokenKey"`,
				`"passwordHash"`,
			},
			ExpectedEvents: map[string]int{
				"OnModelBeforeUpdate":        1,
				"OnModelAfterUpdate":         1,
				"OnAdminBeforeUpdateRequest": 1,
				"OnAdminAfterUpdateRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
		{
			Name:   "OnAdminAfterUpdateRequest error response",
			Method: http.MethodPatch,
			Url:    "/api/admins/2107977127528759298",
			Body: strings.NewReader(`{
				"email":"testnew@example.com",
				"password":"1234567891",
				"passwordConfirm":"1234567891",
				"avatar":5
			}`),
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnAdminAfterUpdateRequest().Add(func(e *core.AdminUpdateEvent) error {
					return errors.New("error")
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			ExpectedEvents: map[string]int{
				"OnModelBeforeUpdate":        1,
				"OnModelAfterUpdate":         1,
				"OnAdminBeforeUpdateRequest": 1,
				"OnAdminAfterUpdateRequest":  1,
			},
			TestAppFactory: func(t *testing.T) *tests.TestApp {
				return app
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(suite.T())
	}
}

type AdminTestSuite struct {
	suite.Suite
	App            *tests.TestApp
	AdminAuthToken string
	UserAuthToken  string
}

func (suite *AdminTestSuite) SetupSuite() {
	app, _ := tests.NewTestApp()
	suite.AdminAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzAyMzYxMTQsImlkIjoiMjEwNzk3NzEyNzUyODc1OTI5NiIsInR5cGUiOiJhZG1pbiJ9.ikCEJR-iPIrZwpbsWjtslMdq75suCAEYfaRK7Oz-NZ0"
	suite.UserAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiIyMTA3OTc3Mzk3MDYzMTIyOTQ0IiwiZXhwIjoxNzMwOTEyMTQzLCJpZCI6Il9wYl91c2Vyc19hdXRoXyIsInR5cGUiOiJhdXRoUmVjb3JkIiwidmVyaWZpZWQiOnRydWV9.Us_731ziRkeeZvYvXiXsc6CKEwdKp4rSvsGbG5L1OUQ"
	suite.App = app
}

func (suite *AdminTestSuite) TearDownSuite() {
	suite.App.Cleanup()
}

func TestAdminTestSuite(t *testing.T) {
	suite.Run(t, new(AdminTestSuite))
}
