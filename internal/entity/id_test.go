package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID_IdTypeAndValue(t *testing.T) {
	assert.Equal(t, ID("type:val").Type(), "type")
	assert.Equal(t, ID("type:val").Id(), "val")

}
