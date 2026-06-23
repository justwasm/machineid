//go:build js && wasm

package machineid

import (
	"syscall/js"
	"uuid"
)

const (
	storageKey = "machineid"
	dbName     = "machineid"
	storeName  = "ids"
	dbVersion  = 1
)

// machineID returns a persistent machine ID stored in IndexedDB.
// If no ID is found, a new random UUID v7 is generated, stored, and returned.
// Falls back to a fresh random ID when IndexedDB is unavailable.
func machineID() (string, error) {
	if id := readIDB(); id != "" {
		return id, nil
	}
	return generateAndStore()
}

func readIDB() string {
	idb := js.Global().Get("indexedDB")
	if !idb.Truthy() {
		return ""
	}
	req := idb.Call("open", dbName, dbVersion)
	c := make(chan string, 1)

	req.Set("onupgradeneeded", js.FuncOf(func(this js.Value, _ []js.Value) any {
		req.Get("result").Call("createObjectStore", storeName)
		return nil
	}))

	req.Set("onsuccess", js.FuncOf(func(this js.Value, _ []js.Value) any {
		db := req.Get("result")
		get := db.Call("transaction", storeName).Call("objectStore", storeName).Call("get", storageKey)

		get.Set("onsuccess", js.FuncOf(func(this js.Value, _ []js.Value) any {
			if v := get.Get("result"); v.Truthy() {
				c <- v.String()
			} else {
				c <- ""
			}
			return nil
		}))
		get.Set("onerror", js.FuncOf(func(this js.Value, _ []js.Value) any {
			c <- ""
			return nil
		}))

		return nil
	}))

	req.Set("onerror", js.FuncOf(func(this js.Value, _ []js.Value) any {
		c <- ""
		return nil
	}))

	return <-c
}

func writeIDB(id string) {
	idb := js.Global().Get("indexedDB")
	if !idb.Truthy() {
		return
	}
	req := idb.Call("open", dbName, dbVersion)
	c := make(chan struct{}, 1)

	req.Set("onupgradeneeded", js.FuncOf(func(this js.Value, _ []js.Value) any {
		req.Get("result").Call("createObjectStore", storeName)
		return nil
	}))

	req.Set("onsuccess", js.FuncOf(func(this js.Value, _ []js.Value) any {
		db := req.Get("result")
		tx := db.Call("transaction", storeName, "readwrite")
		tx.Call("objectStore", storeName).Call("put", id, storageKey)

		tx.Set("oncomplete", js.FuncOf(func(this js.Value, _ []js.Value) any {
			c <- struct{}{}
			return nil
		}))
		tx.Set("onerror", js.FuncOf(func(this js.Value, _ []js.Value) any {
			c <- struct{}{}
			return nil
		}))

		return nil
	}))

	req.Set("onerror", js.FuncOf(func(this js.Value, _ []js.Value) any {
		c <- struct{}{}
		return nil
	}))

	<-c
}

func generateAndStore() (string, error) {
	id := uuid.NewV7().String()
	if js.Global().Get("indexedDB").Truthy() {
		writeIDB(id)
	}
	return id, nil
}
