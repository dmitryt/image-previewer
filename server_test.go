package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("dummy test", func(t *testing.T) {
		require.Equal(t, true, true)
	})
}
