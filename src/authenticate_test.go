package main

import (
	"testing"
	"os"
	"encoding/hex"
	"bytes"
)

func setUpEnvVars() {
	destEnv, userEnv, passEnv := os.Getenv("AUTH_DEST"), os.Getenv("AUTH_USER"), os.Getenv("AUTH_PASS")
	if destEnv != "" && userEnv != "" && passEnv != "" {
		dest, user, pass = destEnv, userEnv, passEnv
	}
}

func TestIntentMessageCorrectlySetsIntentValue(t *testing.T) {
	expected, _ := hex.DecodeString("00000000000000000000010000000a7b000000292300000000001c010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	res := getIntentMessage("7b")

	if !bytes.Equal(expected, res) {
		t.Error("Intent message not as expected")
	}

}
