package joe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	var err error = Error("test") // compiler check to make sure we are actually implementing the "error" interface
	assert.Equal(t, "test", err.Error())
}
