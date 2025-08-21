// genfunc.go
package api

import (
	"time"

	"github.com/gofrs/uuid"
)

func generateNewID() (string, error) {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return newUUID.String(), nil
}

func generateCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
