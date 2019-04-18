package joe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	var err error
	err = Error("test")
	assert.Equal(t, "test", err.Error())
}
