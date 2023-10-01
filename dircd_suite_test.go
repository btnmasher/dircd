package dircd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDircd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dircd Suite")
}
