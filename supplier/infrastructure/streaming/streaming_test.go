package streaming_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestStreaming(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Streaming Suite")
}
