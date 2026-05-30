//go:build js && wasm
// +build js,wasm

package machineid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"syscall/js"
)

const storageKey = "machineid"

// machineID returns a persistent machine ID stored in localStorage.
// If no ID is found, a new random UUID v4 is generated, stored, and returned.
// Falls back to a fresh random ID when localStorage is unavailable.
func machineID() (string, error) {
	if id := readFromStorage(); id != "" {
		return id, nil
	}
	return generateAndStore()
}

func readFromStorage() string {
	storage := js.Global().Get("localStorage")
	if !storage.Truthy() {
		return ""
	}
	v := storage.Call("getItem", storageKey)
	if !v.Truthy() {
		return ""
	}
	return v.String()
}

func generateAndStore() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Format as UUID v4 (RFC 4122).
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	id := fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:]),
	)
	storage := js.Global().Get("localStorage")
	if storage.Truthy() {
		storage.Call("setItem", storageKey, id)
	}
	return id, nil
}
