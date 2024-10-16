package forms_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestForms(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Forms Suite")
}
