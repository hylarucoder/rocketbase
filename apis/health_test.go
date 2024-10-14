package apis_test

import (
	"net/http"
	"testing"

	"github.com/hylarucoder/rocketbase/tests"
	"github.com/stretchr/testify/suite"
)

func (suite *HealthTestSuite) TestHealthAPI() {
	t := suite.T()

	scenarios := []tests.ApiScenario{
		{
			Name:           "health status returns 200",
			Method:         http.MethodGet,
			Url:            "/api/health",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"code":200`,
				`"data":{`,
				`"canBackup":true`,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

type HealthTestSuite struct {
	suite.Suite
	App *tests.TestApp
	Var int
}

func (suite *HealthTestSuite) SetupSuite() {
	app, _ := tests.NewTestApp()
	suite.Var = 5
	suite.App = app
}

func (suite *HealthTestSuite) TearDownSuite() {
	suite.App.Cleanup()
}

func TestHealthTestSuite(t *testing.T) {
	suite.Run(t, new(HealthTestSuite))
}
