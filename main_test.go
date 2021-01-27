package main

import (
	"bufio"
	"bytes"
	"github.com/sonatype-nexus-community/go-sona-types/iq"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_showPolicyActionMessage(t *testing.T) {
	verifyReportURL(t, "anythingElse") //default policy action
	verifyReportURL(t, iq.PolicyActionWarning)
	verifyReportURL(t, iq.PolicyActionFailure)
}

func verifyReportURL(t *testing.T, policyAction string) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)
	theURL := "someURL"
	showPolicyActionMessage(iq.StatusURLResult{AbsoluteReportHTMLURL: theURL, PolicyAction: policyAction}, bufWriter)
	bufWriter.Flush()
	assert.True(t, strings.Contains(buf.String(), "Report URL:  "+theURL), buf.String())
}
