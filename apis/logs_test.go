package apis_test

import (
	"net/http"
	"testing"

	"github.com/hylarucoder/rocketbase/tests"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
)

func (suite *LogsTestSuite) TestLogsList() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodGet,
			Url:             "/api/logs",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "authorized as auth record",
			Method: http.MethodGet,
			Url:    "/api/logs",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "authorized as admin",
			Method: http.MethodGet,
			Url:    "/api/logs",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				if err := tests.MockLogsData(app); err != nil {
					t.Fatal(err)
				}
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalItems":2`,
				`"items":[{`,
				`"id":"873f2133-9f38-44fb-bf82-c8f53b310d91"`,
				`"id":"f2133873-44fb-9f38-bf82-c918f53b310d"`,
			},
		},
		{
			Name:   "authorized as admin + filter",
			Method: http.MethodGet,
			Url:    "/api/logs?filter=data.status>200",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				if err := tests.MockLogsData(app); err != nil {
					t.Fatal(err)
				}
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":1`,
				`"perPage":30`,
				`"totalItems":1`,
				`"items":[{`,
				`"id":"f2133873-44fb-9f38-bf82-c918f53b310d"`,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *LogsTestSuite) TestLogView() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodGet,
			Url:             "/api/logs/873f2133-9f38-44fb-bf82-c8f53b310d91",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "authorized as auth record",
			Method: http.MethodGet,
			Url:    "/api/logs/873f2133-9f38-44fb-bf82-c8f53b310d91",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "authorized as admin (nonexisting request log)",
			Method: http.MethodGet,
			Url:    "/api/logs/missing1-9f38-44fb-bf82-c8f53b310d91",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				if err := tests.MockLogsData(app); err != nil {
					t.Fatal(err)
				}
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "authorized as admin (existing request log)",
			Method: http.MethodGet,
			Url:    "/api/logs/873f2133-9f38-44fb-bf82-c8f53b310d91",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				if err := tests.MockLogsData(app); err != nil {
					t.Fatal(err)
				}
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"id":"873f2133-9f38-44fb-bf82-c8f53b310d91"`,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *LogsTestSuite) TestLogsStats() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:            "unauthorized",
			Method:          http.MethodGet,
			Url:             "/api/logs/stats",
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "authorized as auth record",
			Method: http.MethodGet,
			Url:    "/api/logs/stats",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "authorized as admin",
			Method: http.MethodGet,
			Url:    "/api/logs/stats",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				if err := tests.MockLogsData(app); err != nil {
					t.Fatal(err)
				}
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`[{"total":1,"date":"2022-05-01 10:00:00.000Z"},{"total":1,"date":"2022-05-02 10:00:00.000Z"}]`,
			},
		},
		{
			Name:   "authorized as admin + filter",
			Method: http.MethodGet,
			Url:    "/api/logs/stats?filter=data.status>200",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				if err := tests.MockLogsData(app); err != nil {
					t.Fatal(err)
				}
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`[{"total":1,"date":"2022-05-02 10:00:00.000Z"}]`,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

type LogsTestSuite struct {
	suite.Suite
	App            *tests.TestApp
	AdminAuthToken string
	UserAuthToken  string
}

func (suite *LogsTestSuite) SetupSuite() {
	app, _ := tests.NewTestApp()
	suite.AdminAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzAyMzYxMTQsImlkIjoiMjEwNzk3NzEyNzUyODc1OTI5NiIsInR5cGUiOiJhZG1pbiJ9.ikCEJR-iPIrZwpbsWjtslMdq75suCAEYfaRK7Oz-NZ0"
	suite.UserAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiIyMTA3OTc3Mzk3MDYzMTIyOTQ0IiwiZXhwIjoxNzMwOTEyMTQzLCJpZCI6Il9wYl91c2Vyc19hdXRoXyIsInR5cGUiOiJhdXRoUmVjb3JkIiwidmVyaWZpZWQiOnRydWV9.Us_731ziRkeeZvYvXiXsc6CKEwdKp4rSvsGbG5L1OUQ"
	suite.App = app
}

func (suite *LogsTestSuite) TearDownSuite() {
	suite.App.Cleanup()
}

func TestLogsTestSuite(t *testing.T) {
	suite.Run(t, new(LogsTestSuite))
}
