/*
Package store implements containers and functions to store
data.
*/
package store

import (
	"encoding/json"
	"os"
)

type Store interface {
	// Get is used to retrieve data from the implementation store.
	Get(id string) interface{}

	// Set is used to store data using the provided id as its key.
	Set(id string, data interface{})
}

// Implementation of Store that is used to store data in JSON format
// in the file system.
type JSONFileStore[T interface{}] struct {
	Path string
}

// Load reads the file content from the path specified and creates
// a JSON representation of the content using the provided type T.
func (jfs JSONFileStore[T]) Load() (*T, error) {
	rawJSON, err := os.ReadFile(jfs.Path)
	if err != nil {
		// Failed to load JSON.
		return nil, err
	}

	var parsedJSON T
	err = json.Unmarshal(rawJSON, &parsedJSON)
	if err != nil {
		// Failed to convert JSON string to Config struct.
		return nil, err
	}

	return &parsedJSON, nil
}

func (jfs JSONFileStore[T]) Save(data T) error {
	rawJSON, err := json.Marshal(data)
	if err != nil {
		// Failed to stringify JSON.
		return err
	}

	os.WriteFile(jfs.Path, rawJSON, os.ModePerm)

	return nil
}
