package entity

import (
	"fmt"
	"strings"
)

type ID string

func (id ID) Type() string {
	return strings.Split(string(id), ":")[0]
}
func (id ID) Id() string {
	return strings.Split(string(id), ":")[1]

}

func (id ID) String() string {
	return string(id)
}

func NewID[T any](entityType string, id T) ID {
	return ID(fmt.Sprintf("%s:%v", entityType, id))
}
