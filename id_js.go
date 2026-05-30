//go:build js && wasm
// +build js,wasm

package machineid

import (
	"syscall/js"

	"github.com/google/uuid"
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
	id := uuid.New().String()
	storage := js.Global().Get("localStorage")
	if storage.Truthy() {
		storage.Call("setItem", storageKey, id)
	}
	return id, nil
}
