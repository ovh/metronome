package factories

import (
	"errors"

	"github.com/ovh/metronome/src/api/models"
	log "github.com/sirupsen/logrus"
)

// Error create a models.Error from an error
func Error(err error) models.Error {
	if err == nil {
		err = errors.New("")
		log.Warn("Nil error connot be formatted as models.Error")
	}

	return models.Error{
		Err: err.Error(),
	}
}
