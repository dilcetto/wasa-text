// genfunc.go
package api

import "github.com/gofrs/uuid"

func generateNewID() (string, error) {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return newUUID.String(), nil
}
