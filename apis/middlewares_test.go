package apis_test

import (
	"net/http"
	"testing"

	"github.com/hylarucoder/rocketbase/apis"
	"github.com/hylarucoder/rocketbase/tests"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
)

func (suite *MiddlewaresTestSuite) TestRequireGuestOnly() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:   "valid record token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireGuestOnly(),
					},
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid admin token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireGuestOnly(),
					},
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "expired/invalid token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken + "1",
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireGuestOnly(),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
		{
			Name:   "guest",
			Method: http.MethodGet,
			Url:    "/my/test",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireGuestOnly(),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *MiddlewaresTestSuite) TestRequireRecordAuth() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:   "guest",
			Method: http.MethodGet,
			Url:    "/my/test",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireRecordAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "expired/invalid token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken + "1",
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireRecordAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid admin token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireRecordAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireRecordAuth(),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
		{
			Name:   "valid record token with collection not in the restricted list",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireRecordAuth("demo1", "demo2"),
					},
				})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token with collection in the restricted list",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireRecordAuth("demo1", "demo2", "users"),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *MiddlewaresTestSuite) TestRequireSameContextRecordAuth() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:   "guest",
			Method: http.MethodGet,
			Url:    "/my/users/test",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/:collection/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireSameContextRecordAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "expired/invalid token",
			Method: http.MethodGet,
			Url:    "/my/users/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken + "1",
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/:collection/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireSameContextRecordAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid admin token",
			Method: http.MethodGet,
			Url:    "/my/users/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/:collection/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireSameContextRecordAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token but from different collection",
			Method: http.MethodGet,
			Url:    "/my/users/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/:collection/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireSameContextRecordAuth(),
					},
				})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireRecordAuth(),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *MiddlewaresTestSuite) TestRequireAdminAuth() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:   "guest",
			Method: http.MethodGet,
			Url:    "/my/test",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "expired/invalid token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken + "1",
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid admin token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminAuth(),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *MiddlewaresTestSuite) TestRequireAdminAuthOnlyIfAny() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:   "guest (while having at least 1 existing admin)",
			Method: http.MethodGet,
			Url:    "/my/test",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminAuthOnlyIfAny(app),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "guest (while having 0 existing admins)",
			Method: http.MethodGet,
			Url:    "/my/test",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				// delete all admins
				_, err := app.Dao().DB().NewQuery("DELETE FROM {{_admins}}").Execute()
				if err != nil {
					t.Fatal(err)
				}

				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminAuthOnlyIfAny(app),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
		{
			Name:   "expired/invalid token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken + "1",
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminAuthOnlyIfAny(app),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminAuthOnlyIfAny(app),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid admin token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminAuthOnlyIfAny(app),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *MiddlewaresTestSuite) TestRequireAdminOrRecordAuth() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:   "guest",
			Method: http.MethodGet,
			Url:    "/my/test",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrRecordAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "expired/invalid token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken + "1",
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrRecordAuth(),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrRecordAuth(),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
		{
			Name:   "valid record token with collection not in the restricted list",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrRecordAuth("demo1", "demo2", "clients"),
					},
				})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token with collection in the restricted list",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrRecordAuth("demo1", "demo2", "users"),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
		{
			Name:   "valid admin token",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrRecordAuth(),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
		{
			Name:   "valid admin token + restricted collections list (should be ignored)",
			Method: http.MethodGet,
			Url:    "/my/test",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrRecordAuth("demo1", "demo2"),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *MiddlewaresTestSuite) TestRequireAdminOrOwnerAuth() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:   "guest",
			Method: http.MethodGet,
			Url:    "/my/test/2107977397063122944",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test/:id",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrOwnerAuth(""),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "expired/invalid token",
			Method: http.MethodGet,
			Url:    "/my/test/2107977397063122944",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken + "1",
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test/:id",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrOwnerAuth(""),
					},
				})
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token (different user)",
			Method: http.MethodGet,
			Url:    "/my/test/2107977397063122944",
			RequestHeaders: map[string]string{
				// TODO: use2
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test/:id",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrOwnerAuth(""),
					},
				})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token (different collection)",
			Method: http.MethodGet,
			Url:    "/my/test/2107977397063122944",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test/:id",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrOwnerAuth(""),
					},
				})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "valid record token (owner)",
			Method: http.MethodGet,
			Url:    "/my/test/2107977397063122944",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test/:id",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrOwnerAuth(""),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
		{
			Name:   "valid admin token",
			Method: http.MethodGet,
			Url:    "/my/test/2107977397063122944",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/test/:custom",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.RequireAdminOrOwnerAuth("custom"),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func (suite *MiddlewaresTestSuite) TestLoadCollectionContext() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:   "missing collection",
			Method: http.MethodGet,
			Url:    "/my/missing",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/:collection",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.LoadCollectionContext(app),
					},
				})
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "guest",
			Method: http.MethodGet,
			Url:    "/my/demo1",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/:collection",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.LoadCollectionContext(app),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
		{
			Name:   "valid record token",
			Method: http.MethodGet,
			Url:    "/my/demo1",
			RequestHeaders: map[string]string{
				"Authorization": suite.UserAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/:collection",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.LoadCollectionContext(app),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
		{
			Name:   "valid admin token",
			Method: http.MethodGet,
			Url:    "/my/demo1",
			RequestHeaders: map[string]string{
				"Authorization": suite.AdminAuthToken,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/:collection",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.LoadCollectionContext(app),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
		{
			Name:   "mismatched type",
			Method: http.MethodGet,
			Url:    "/my/demo1",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/:collection",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.LoadCollectionContext(app, "auth"),
					},
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "matched type",
			Method: http.MethodGet,
			Url:    "/my/users",
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				e.AddRoute(echo.Route{
					Method: http.MethodGet,
					Path:   "/my/:collection",
					Handler: func(c echo.Context) error {
						return c.String(200, "test123")
					},
					Middlewares: []echo.MiddlewareFunc{
						apis.LoadCollectionContext(app, "auth"),
					},
				})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"test123"},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

type MiddlewaresTestSuite struct {
	suite.Suite
	App            *tests.TestApp
	AdminAuthToken string
	UserAuthToken  string
}

func (suite *MiddlewaresTestSuite) SetupSuite() {
	app, _ := tests.NewTestApp()
	suite.AdminAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzAyMzYxMTQsImlkIjoiMjEwNzk3NzEyNzUyODc1OTI5NiIsInR5cGUiOiJhZG1pbiJ9.ikCEJR-iPIrZwpbsWjtslMdq75suCAEYfaRK7Oz-NZ0"
	suite.UserAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiIyMTA3OTc3Mzk3MDYzMTIyOTQ0IiwiZXhwIjoxNzMwOTEyMTQzLCJpZCI6Il9wYl91c2Vyc19hdXRoXyIsInR5cGUiOiJhdXRoUmVjb3JkIiwidmVyaWZpZWQiOnRydWV9.Us_731ziRkeeZvYvXiXsc6CKEwdKp4rSvsGbG5L1OUQ"
	suite.App = app
}

func (suite *MiddlewaresTestSuite) TearDownSuite() {
	suite.App.Cleanup()
}

func TestMiddlewaresTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewaresTestSuite))
}
