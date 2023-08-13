package localfiles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsLocal(t *testing.T) {

	assert.False(t, false, IsLocal("test.dat"))
}
