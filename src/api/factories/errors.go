package factories

import "github.com/ovh/metronome/src/api/models"

// Error create a models.Error from an error
func Error(err error) models.Error {
	return models.Error{
		Err: err.Error(),
	}
}
