package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccount(t *testing.T) {
	a := &Account{}

	require.NoError(t, a.SetPassword("abcdefg"))
	require.Equal(t, true, a.ComparePassword("abcdefg"))
	require.Equal(t, false, a.ComparePassword("abcdef"))
}
