package pocketbase_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRocketbase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rocketbase Suite")
}
