package gwtypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccessMode(t *testing.T) {
	tests := []struct {
		mode AccessMode
		exp  string
	}{
		{mode: Input, exp: "INPUT"},
		{mode: Output, exp: "OUTPUT"},
		{mode: Append, exp: "APPEND"},
		{mode: Random, exp: "RANDOM"},
	}

	for _, tt := range tests {
		assert.EqualValues(t, tt.exp, tt.mode.String())
	}
}

func TestLockMode(t *testing.T) {
	tests := []struct {
		mode LockMode
		exp  string
	}{
		{mode: Shared, exp: "SHARED"},
		{mode: LockRead, exp: "LOCK READ"},
		{mode: LockWrite, exp: "LOCK WRITE"},
		{mode: LockReadWrite, exp: "LOCK READ WRITE"},
		{mode: Default, exp: "DEFAULT"},
	}

	for _, tt := range tests {
		assert.EqualValues(t, tt.exp, tt.mode.String())
	}
}
