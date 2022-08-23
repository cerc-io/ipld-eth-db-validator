package validator_test

import (
	"io/ioutil"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestETHSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "eth ipld validator eth suite test")
}

var _ = BeforeSuite(func() {
	logrus.SetOutput(ioutil.Discard)
})
