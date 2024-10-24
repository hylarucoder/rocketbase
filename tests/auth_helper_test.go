package tests

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hylarucoder/rocketbase/tools/security"
)

func TestRecordAuth(t *testing.T) {
	app, _ := NewTestApp()
	defer app.Cleanup()
	//record, _ := app.Dao().FindFirstRecordByFilter(""+
	//	"users",
	//	`email = 'test@example.com'`,
	//)
	//record, _ := app.Dao().FindFirstRecordByFilter(""+
	//	"clients",
	//	`email = 'test@example.com'`,
	//)
	//token, _ := security.NewJWT(
	//	jwt.MapClaims{
	//		"id":           record.Collection().Id,
	//		"type":         "authRecord",
	//		"collectionId": record.Id,
	//		"verified":     true,
	//	},
	//	record.TokenKey()+app.Settings().RecordAuthToken.Secret,
	//	app.Settings().RecordAuthToken.Duration,
	//)
	record, _ := app.Dao().FindAdminByEmail("test@example.com")
	token, _ := security.NewJWT(
		jwt.MapClaims{
			"id":   record.Id,
			"type": "admin",
		},
		record.TokenKey+app.Settings().AdminAuthToken.Secret,
		app.Settings().AdminAuthToken.Duration,
	)
	println("token =>", token)
}
