package httpapi

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHttpapi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Httpapi Suite")
}
