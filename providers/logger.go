package providers

import (
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

// SetLogger is used to pass the loger from the main program
func SetLogger(logger *logrus.Entry) { log = logger }
