package forms_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hylarucoder/rocketbase/forms"
	"github.com/hylarucoder/rocketbase/tests"
	"github.com/stretchr/testify/suite"
)

func (suite *S3FilesystemTestSuite) TestS3FilesystemValidate() {
	t := suite.T()

	app := suite.App

	scenarios := []struct {
		name           string
		filesystem     string
		expectedErrors []string
	}{
		{
			"empty filesystem",
			"",
			[]string{"filesystem"},
		},
		{
			"invalid filesystem",
			"something",
			[]string{"filesystem"},
		},
		{
			"backups filesystem",
			"backups",
			[]string{},
		},
		{
			"storage filesystem",
			"storage",
			[]string{},
		},
	}

	for _, s := range scenarios {
		form := forms.NewTestS3Filesystem(app)
		form.Filesystem = s.filesystem

		result := form.Validate()

		// parse errors
		errs, ok := result.(validation.Errors)
		if !ok && result != nil {
			t.Errorf("[%s] Failed to parse errors %v", s.name, result)
			continue
		}

		// check errors
		if len(errs) > len(s.expectedErrors) {
			t.Errorf("[%s] Expected error keys %v, got %v", s.name, s.expectedErrors, errs)
			continue
		}
		for _, k := range s.expectedErrors {
			if _, ok := errs[k]; !ok {
				t.Errorf("[%s] Missing expected error key %q in %v", s.name, k, errs)
			}
		}
	}
}

func (suite *S3FilesystemTestSuite) TestS3FilesystemSubmitFailure() {
	t := suite.T()

	app := suite.App

	// check if validate was called
	{
		form := forms.NewTestS3Filesystem(app)
		form.Filesystem = ""

		result := form.Submit()

		if result == nil {
			t.Fatal("Expected error, got nil")
		}

		if _, ok := result.(validation.Errors); !ok {
			t.Fatalf("Expected validation.Error, got %v", result)
		}
	}

	// check with valid storage and disabled s3
	{
		form := forms.NewTestS3Filesystem(app)
		form.Filesystem = "storage"

		result := form.Submit()

		if result == nil {
			t.Fatal("Expected error, got nil")
		}

		if _, ok := result.(validation.Error); ok {
			t.Fatalf("Didn't expect validation.Error, got %v", result)
		}
	}
}

type S3FilesystemTestSuite struct {
	suite.Suite
	App *tests.TestApp
	Var int
}

func (suite *S3FilesystemTestSuite) SetupSuite() {
	app, _ := tests.NewTestApp()
	suite.Var = 5
	suite.App = app
}

func (suite *S3FilesystemTestSuite) TearDownSuite() {
	suite.App.Cleanup()
}

func TestS3FilesystemTestSuite(t *testing.T) {
	suite.Run(t, new(S3FilesystemTestSuite))
}
