package cmd

import (
	"bytes"
	"log"
	"testing"
)

func TestVerboseLog_Enabled(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	old := verboseMode
	verboseMode = true
	defer func() { verboseMode = old }()

	verboseLog("test %s", "message")

	if !bytes.Contains(buf.Bytes(), []byte("[verbose] test message")) {
		t.Errorf("expected verbose log output, got: %s", buf.String())
	}
}

func TestVerboseLog_Disabled(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	old := verboseMode
	verboseMode = false
	defer func() { verboseMode = old }()

	verboseLog("should not appear")

	if buf.Len() != 0 {
		t.Errorf("expected no output when verbose disabled, got: %s", buf.String())
	}
}
