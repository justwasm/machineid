//go:build darwin

package machineid

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// machineID returns the uuid returned by `ioreg -rd1 -c IOPlatformExpertDevice`.
// If there is an error running the commad an empty string is returned.
func machineID() (string, error) {
	buf := &bytes.Buffer{}
	err := run(buf, os.Stderr, "ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	if err != nil {
		return "", err
	}
	id, err := extractID(buf.String())
	if err != nil {
		return "", err
	}
	return trim(id), nil
}

func extractID(lines string) (string, error) {
	for line := range strings.SplitSeq(lines, "\n") {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.SplitAfter(line, `" = "`)
			if len(parts) == 2 {
				return strings.TrimRight(parts[1], `"`), nil
			}
		}
	}
	return "", fmt.Errorf("failed to extract 'IOPlatformUUID' value from `ioreg` output.\n%s", lines)
}
