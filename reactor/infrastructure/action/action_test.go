package action_test

import (
	"fmt"
	"io"
	"reflect"
	"testing"
	"testing/iotest"

	"github.com/chitoku-k/ejaculation-counter/reactor/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

func TestAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Action Suite")
}

type ReplyEventEqualMatcher struct {
	Expected service.ReplyEvent
}

func (matcher *ReplyEventEqualMatcher) Match(actual interface{}) (success bool, err error) {
	actualReplyEvent, ok := actual.(service.ReplyEvent)
	if !ok {
		return false, fmt.Errorf("ReplyEventEqual matcher expects a reply event.  Got:\n%s", format.Object(actual, 1))
	}

	expectedBody, err := io.ReadAll(matcher.Expected.Body)
	if err != nil {
		return false, fmt.Errorf("ReplyEventEqual matcher could not read body.  Got:\n%s", err)
	}

	err = iotest.TestReader(actualReplyEvent.Body, expectedBody)
	if err != nil {
		return false, err
	}

	actualReplyEvent.Body = matcher.Expected.Body
	return reflect.DeepEqual(actualReplyEvent, matcher.Expected), nil
}

func (matcher *ReplyEventEqualMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to equal", matcher.Expected)
}

func (matcher *ReplyEventEqualMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to equal", matcher.Expected)
}

func ReplyEventEqual(expected service.ReplyEvent) types.GomegaMatcher {
	return &ReplyEventEqualMatcher{
		Expected: expected,
	}
}
