package tokens_test

import (
	"fmt"
	"testing"

	"github.com/hylarucoder/rocketbase/tests"
	"github.com/hylarucoder/rocketbase/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func (suite *RecordTestSuite) TestNewRecordAuthToken() {
	app := suite.App
	assert.NotNil(suite.T(), app)

	user, err := app.Dao().FindAuthRecordByEmail("users", "test@example.com")
	assert.Nil(suite.T(), err)

	token, err := tokens.NewRecordAuthToken(app, user)
	assert.Nil(suite.T(), err)

	tokenRecord, _ := app.Dao().FindAuthRecordByToken(
		token,
		app.Settings().RecordAuthToken.Secret,
	)
	assert.Equal(suite.T(), user.Id, tokenRecord.Id)
}

func (suite *RecordTestSuite) TestNewRecordVerifyToken() {
	app := suite.App

	user, err := app.Dao().FindAuthRecordByEmail("users", "test@example.com")
	assert.Nil(suite.T(), err)

	token, err := tokens.NewRecordVerifyToken(app, user)
	assert.Nil(suite.T(), err)

	tokenRecord, _ := app.Dao().FindAuthRecordByToken(
		token,
		app.Settings().RecordVerificationToken.Secret,
	)
	assert.Equal(suite.T(), user.Id, tokenRecord.Id)
}

func (suite *RecordTestSuite) TestNewRecordResetPasswordToken() {
	app := suite.App

	user, err := app.Dao().FindAuthRecordByEmail("users", "test@example.com")
	assert.Nil(suite.T(), err)

	token, err := tokens.NewRecordResetPasswordToken(app, user)
	assert.Nil(suite.T(), err)

	tokenRecord, _ := app.Dao().FindAuthRecordByToken(
		token,
		app.Settings().RecordPasswordResetToken.Secret,
	)
	assert.Equal(suite.T(), user.Id, tokenRecord.Id)
}

func (suite *RecordTestSuite) TestNewRecordChangeEmailToken() {
	app := suite.App
	user, err := app.Dao().FindAuthRecordByEmail("users", "test@example.com")
	assert.Nil(suite.T(), err)

	token, err := tokens.NewRecordChangeEmailToken(app, user, "test_new@example.com")
	assert.Nil(suite.T(), err)

	tokenRecord, _ := app.Dao().FindAuthRecordByToken(
		token,
		app.Settings().RecordEmailChangeToken.Secret,
	)
	assert.Equal(suite.T(), user.Id, tokenRecord.Id)
}

func (suite *RecordTestSuite) TestNewRecordFileToken() {
	app := suite.App

	user, err := app.Dao().FindAuthRecordByEmail("users", "test@example.com")
	assert.Nil(suite.T(), err)

	token, err := tokens.NewRecordFileToken(app, user)
	assert.Nil(suite.T(), err)

	tokenRecord, _ := app.Dao().FindAuthRecordByToken(
		token,
		app.Settings().RecordFileToken.Secret,
	)
	assert.Equal(suite.T(), user.Id, tokenRecord.Id)
}

type RecordTestSuite struct {
	suite.Suite
	App *tests.TestApp
	Var int
}

func (suite *RecordTestSuite) SetupTest() {
	app, _ := tests.NewTestApp()
	suite.Var = 5
	suite.App = app
}

func (suite *RecordTestSuite) TearDownTest() {
	suite.App.Cleanup()
}

func (suite *RecordTestSuite) SetupSuite() {
	fmt.Println("setup suite")
}

func (suite *RecordTestSuite) TearDownSuite() {
	fmt.Println("teardown suite")
}

func TestRecordTestSuite(t *testing.T) {
	suite.Run(t, new(RecordTestSuite))
}
