package apis_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hylarucoder/rocketbase/apis"
	"github.com/hylarucoder/rocketbase/core"
	"github.com/hylarucoder/rocketbase/daos"
	"github.com/hylarucoder/rocketbase/models"
	"github.com/hylarucoder/rocketbase/tests"
	"github.com/hylarucoder/rocketbase/tools/hook"
	"github.com/hylarucoder/rocketbase/tools/subscriptions"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/suite"
)

func (suite *RealtimeTestSuite) TestRealtimeConnect() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Method:         http.MethodGet,
			Url:            "/api/realtime",
			Timeout:        100 * time.Millisecond,
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`id:`,
				`event:PB_CONNECT`,
				`data:{"clientId":`,
			},
			ExpectedEvents: map[string]int{
				"OnRealtimeConnectRequest":    1,
				"OnRealtimeBeforeMessageSend": 1,
				"OnRealtimeAfterMessageSend":  1,
				"OnRealtimeDisconnectRequest": 1,
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				if len(app.SubscriptionsBroker().Clients()) != 0 {
					t.Errorf("Expected the subscribers to be removed after connection close, found %d", len(app.SubscriptionsBroker().Clients()))
				}
			},
		},
		{
			Name:           "PB_CONNECT interrupt",
			Method:         http.MethodGet,
			Url:            "/api/realtime",
			Timeout:        100 * time.Millisecond,
			ExpectedStatus: 200,
			ExpectedEvents: map[string]int{
				"OnRealtimeConnectRequest":    1,
				"OnRealtimeBeforeMessageSend": 1,
				"OnRealtimeDisconnectRequest": 1,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnRealtimeBeforeMessageSend().Add(func(e *core.RealtimeMessageEvent) error {
					if e.Message.Name == "PB_CONNECT" {
						return errors.New("PB_CONNECT error")
					}
					return nil
				})
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				if len(app.SubscriptionsBroker().Clients()) != 0 {
					t.Errorf("Expected the subscribers to be removed after connection close, found %d", len(app.SubscriptionsBroker().Clients()))
				}
			},
		},
		{
			Name:           "Skipping/ignoring messages",
			Method:         http.MethodGet,
			Url:            "/api/realtime",
			Timeout:        100 * time.Millisecond,
			ExpectedStatus: 200,
			ExpectedEvents: map[string]int{
				"OnRealtimeConnectRequest":    1,
				"OnRealtimeBeforeMessageSend": 1,
				"OnRealtimeDisconnectRequest": 1,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.OnRealtimeBeforeMessageSend().Add(func(e *core.RealtimeMessageEvent) error {
					return hook.StopPropagation
				})
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				if len(app.SubscriptionsBroker().Clients()) != 0 {
					t.Errorf("Expected the subscribers to be removed after connection close, found %d", len(app.SubscriptionsBroker().Clients()))
				}
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t, nil)
	}
}

func (suite *RealtimeTestSuite) TestRealtimeSubscribe() {
	client := subscriptions.NewDefaultClient()

	resetClient := func() {
		client.Unsubscribe()
		client.Set(apis.ContextAdminKey, nil)
		client.Set(apis.ContextAuthRecordKey, nil)
	}

	scenarios := []tests.ApiScenario{
		{
			Name:            "missing client",
			Method:          http.MethodPost,
			Url:             "/api/realtime",
			Body:            strings.NewReader(`{"clientId":"missing","subscriptions":["test1", "test2"]}`),
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:           "existing client - empty subscriptions",
			Method:         http.MethodPost,
			Url:            "/api/realtime",
			Body:           strings.NewReader(`{"clientId":"` + client.Id() + `","subscriptions":[]}`),
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnRealtimeBeforeSubscribeRequest": 1,
				"OnRealtimeAfterSubscribeRequest":  1,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				client.Subscribe("test0")
				app.SubscriptionsBroker().Register(client)
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				if len(client.Subscriptions()) != 0 {
					t.Errorf("Expected no subscriptions, got %v", client.Subscriptions())
				}
				resetClient()
			},
		},
		{
			Name:           "existing client - 2 new subscriptions",
			Method:         http.MethodPost,
			Url:            "/api/realtime",
			Body:           strings.NewReader(`{"clientId":"` + client.Id() + `","subscriptions":["test1", "test2"]}`),
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnRealtimeBeforeSubscribeRequest": 1,
				"OnRealtimeAfterSubscribeRequest":  1,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				client.Subscribe("test0")
				app.SubscriptionsBroker().Register(client)
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				expectedSubs := []string{"test1", "test2"}
				if len(expectedSubs) != len(client.Subscriptions()) {
					t.Errorf("Expected subscriptions %v, got %v", expectedSubs, client.Subscriptions())
				}

				for _, s := range expectedSubs {
					if !client.HasSubscription(s) {
						t.Errorf("Cannot find %q subscription in %v", s, client.Subscriptions())
					}
				}
				resetClient()
			},
		},
		{
			Name:   "existing client - authorized admin",
			Method: http.MethodPost,
			Url:    "/api/realtime",
			Body:   strings.NewReader(`{"clientId":"` + client.Id() + `","subscriptions":["test1", "test2"]}`),
			RequestHeaders: map[string]string{
				"Authorization": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhZG1pbiIsImV4cCI6MjIwODk4NTI2MX0.M1m--VOqGyv0d23eeUc0r9xE8ZzHaYVmVFw1VZW6gT8",
			},
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnRealtimeBeforeSubscribeRequest": 1,
				"OnRealtimeAfterSubscribeRequest":  1,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.SubscriptionsBroker().Register(client)
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				admin, _ := client.Get(apis.ContextAdminKey).(*models.Admin)
				if admin == nil {
					t.Errorf("Expected admin auth model, got nil")
				}
				resetClient()
			},
		},
		{
			Name:   "existing client - authorized record",
			Method: http.MethodPost,
			Url:    "/api/realtime",
			Body:   strings.NewReader(`{"clientId":"` + client.Id() + `","subscriptions":["test1", "test2"]}`),
			RequestHeaders: map[string]string{
				"Authorization": "eyJhbGciOiJIUzI1NiJ9.eyJpZCI6IjRxMXhsY2xtZmxva3UzMyIsInR5cGUiOiJhdXRoUmVjb3JkIiwiY29sbGVjdGlvbklkIjoiX3BiX3VzZXJzX2F1dGhfIiwiZXhwIjoyMjA4OTg1MjYxfQ.UwD8JvkbQtXpymT09d7J6fdA0aP9g4FJ1GPh_ggEkzc",
			},
			ExpectedStatus: 204,
			ExpectedEvents: map[string]int{
				"OnRealtimeBeforeSubscribeRequest": 1,
				"OnRealtimeAfterSubscribeRequest":  1,
			},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				app.SubscriptionsBroker().Register(client)
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				authRecord, _ := client.Get(apis.ContextAuthRecordKey).(*models.Record)
				if authRecord == nil {
					t.Errorf("Expected auth record model, got nil")
				}
				resetClient()
			},
		},
		{
			Name:   "existing client - mismatched auth",
			Method: http.MethodPost,
			Url:    "/api/realtime",
			Body:   strings.NewReader(`{"clientId":"` + client.Id() + `","subscriptions":["test1", "test2"]}`),
			RequestHeaders: map[string]string{
				"Authorization": "eyJhbGciOiJIUzI1NiJ9.eyJpZCI6IjRxMXhsY2xtZmxva3UzMyIsInR5cGUiOiJhdXRoUmVjb3JkIiwiY29sbGVjdGlvbklkIjoiX3BiX3VzZXJzX2F1dGhfIiwiZXhwIjoyMjA4OTg1MjYxfQ.UwD8JvkbQtXpymT09d7J6fdA0aP9g4FJ1GPh_ggEkzc",
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			BeforeTestFunc: func(t *testing.T, app *tests.TestApp, e *echo.Echo) {
				initialAuth := &models.Record{}
				initialAuth.RefreshId()
				client.Set(apis.ContextAuthRecordKey, initialAuth)

				app.SubscriptionsBroker().Register(client)
			},
			AfterTestFunc: func(t *testing.T, app *tests.TestApp, res *http.Response) {
				authRecord, _ := client.Get(apis.ContextAuthRecordKey).(*models.Record)
				if authRecord == nil {
					t.Errorf("Expected auth record model, got nil")
				}
				resetClient()
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t, nil)
	}
}

func (suite *RealtimeTestSuite) TestRealtimeAuthRecordDeleteEvent() {
	t := suite.T()
	app := suite.App
	apis.InitApi(app)

	authRecord, err := app.Dao().FindFirstRecordByData("users", "email", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	client := subscriptions.NewDefaultClient()
	client.Set(apis.ContextAuthRecordKey, authRecord)
	app.SubscriptionsBroker().Register(client)

	e := new(core.ModelEvent)
	e.Dao = app.Dao()
	e.Model = authRecord
	app.OnModelAfterDelete().Trigger(e)

	if len(app.SubscriptionsBroker().Clients()) != 0 {
		t.Fatalf("Expected no subscription clients, found %d", len(app.SubscriptionsBroker().Clients()))
	}
}

func (suite *RealtimeTestSuite) TestRealtimeAuthRecordUpdateEvent() {
	t := suite.T()
	app := suite.App

	apis.InitApi(app)

	authRecord1, err := app.Dao().FindFirstRecordByData("users", "email", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	client := subscriptions.NewDefaultClient()
	client.Set(apis.ContextAuthRecordKey, authRecord1)
	app.SubscriptionsBroker().Register(client)

	// refetch the authRecord and change its email
	authRecord2, err := app.Dao().FindFirstRecordByData("users", "email", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}
	authRecord2.SetEmail("new@example.com")

	e := new(core.ModelEvent)
	e.Dao = app.Dao()
	e.Model = authRecord2
	app.OnModelAfterUpdate().Trigger(e)

	clientAuthRecord, _ := client.Get(apis.ContextAuthRecordKey).(*models.Record)
	if clientAuthRecord.Email() != authRecord2.Email() {
		t.Fatalf("Expected authRecord with email %q, got %q", authRecord2.Email(), clientAuthRecord.Email())
	}
}

func (suite *RealtimeTestSuite) TestRealtimeAdminDeleteEvent() {
	t := suite.T()
	app := suite.App

	apis.InitApi(app)

	admin, err := app.Dao().FindAdminByEmail("test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	client := subscriptions.NewDefaultClient()
	client.Set(apis.ContextAdminKey, admin)
	app.SubscriptionsBroker().Register(client)

	e := new(core.ModelEvent)
	e.Dao = app.Dao()
	e.Model = admin
	app.OnModelAfterDelete().Trigger(e)

	if len(app.SubscriptionsBroker().Clients()) != 0 {
		t.Fatalf("Expected no subscription clients, found %d", len(testApp.SubscriptionsBroker().Clients()))
	}
}

func (suite *RealtimeTestSuite) TestRealtimeAdminUpdateEvent() {
	t := suite.T()
	app := suite.App

	apis.InitApi(app)

	admin1, err := app.Dao().FindAdminByEmail("test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	client := subscriptions.NewDefaultClient()
	client.Set(apis.ContextAdminKey, admin1)
	app.SubscriptionsBroker().Register(client)

	// refetch the authRecord and change its email
	admin2, err := app.Dao().FindAdminByEmail("test@example.com")
	if err != nil {
		t.Fatal(err)
	}
	admin2.Email = "new@example.com"

	e := new(core.ModelEvent)
	e.Dao = app.Dao()
	e.Model = admin2
	app.OnModelAfterUpdate().Trigger(e)

	clientAdmin, _ := client.Get(apis.ContextAdminKey).(*models.Admin)
	if clientAdmin.Email != admin2.Email {
		t.Fatalf("Expected authRecord with email %q, got %q", admin2.Email, clientAdmin.Email)
	}
}

// Custom auth record model struct
// -------------------------------------------------------------------
var _ models.Model = (*CustomUser)(nil)

type CustomUser struct {
	models.BaseModel

	Email string `db:"email" json:"email"`
}

func (m *CustomUser) TableName() string {
	return "users" // the name of your collection
}

func findCustomUserByEmail(dao *daos.Dao, email string) (*CustomUser, error) {
	model := &CustomUser{}

	err := dao.ModelQuery(model).
		AndWhere(dbx.HashExp{"email": email}).
		Limit(1).
		One(model)

	if err != nil {
		return nil, err
	}

	return model, nil
}

func (suite *RealtimeTestSuite) TestRealtimeCustomAuthModelDeleteEvent() {
	t := suite.T()
	app := suite.App

	apis.InitApi(app)

	authRecord, err := app.Dao().FindFirstRecordByData("users", "email", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	client := subscriptions.NewDefaultClient()
	client.Set(apis.ContextAuthRecordKey, authRecord)
	app.SubscriptionsBroker().Register(client)

	// refetch the authRecord as CustomUser
	customUser, err := findCustomUserByEmail(app.Dao(), "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	// delete the custom user (should unset the client auth record)
	if err := app.Dao().Delete(customUser); err != nil {
		t.Fatal(err)
	}

	if len(app.SubscriptionsBroker().Clients()) != 0 {
		t.Fatalf("Expected no subscription clients, found %d", len(app.SubscriptionsBroker().Clients()))
	}
}

func (suite *RealtimeTestSuite) TestRealtimeCustomAuthModelUpdateEvent() {
	t := suite.T()
	app := suite.App

	apis.InitApi(app)

	authRecord, err := app.Dao().FindFirstRecordByData("users", "email", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	client := subscriptions.NewDefaultClient()
	client.Set(apis.ContextAuthRecordKey, authRecord)
	app.SubscriptionsBroker().Register(client)

	// refetch the authRecord as CustomUser
	customUser, err := findCustomUserByEmail(app.Dao(), "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	// change its email
	customUser.Email = "new@example.com"
	if err := app.Dao().Save(customUser); err != nil {
		t.Fatal(err)
	}

	clientAuthRecord, _ := client.Get(apis.ContextAuthRecordKey).(*models.Record)
	if clientAuthRecord.Email() != customUser.Email {
		t.Fatalf("Expected authRecord with email %q, got %q", customUser.Email, clientAuthRecord.Email())
	}
}

type RealtimeTestSuite struct {
	suite.Suite
	App *tests.TestApp
	Var int
}

func (suite *RealtimeTestSuite) SetupSuite() {
	app, _ := tests.NewTestApp()
	suite.Var = 5
	suite.App = app
}

func (suite *RealtimeTestSuite) TearDownSuite() {
	suite.App.Cleanup()
}

func TestRealtimeTestSuite(t *testing.T) {
	suite.Run(t, new(RealtimeTestSuite))
}
