package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	u := &User{}

	require.NoError(t, u.SetPassword("abcdefg"))
	require.Equal(t, true, u.ComparePassword("abcdefg"))
	require.Equal(t, false, u.ComparePassword("abcdef"))
}
