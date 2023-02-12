package validator_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestETHSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "eth ipld validator eth suite test")
}
