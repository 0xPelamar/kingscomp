package jsonhelper

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

func Encode[T any](t T) []byte {
	b, err := json.Marshal(t)
	if err != nil {
		logrus.WithError(err).Fatalln("failed to encode entity")
		return nil
	}
	return b
}

func Decode[T any](b []byte) T {
	var t T
	err := json.Unmarshal(b, &t)
	if err != nil {
		logrus.WithError(err).WithField("val", string(b)).Fatalln("failed to decode entity")
	}
	return t
}
