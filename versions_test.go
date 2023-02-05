package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemanticVersion(t *testing.T) {
	sv, err := NewSemanticVersion("v1.0.0")
	require.NoError(t, err)

	assert.False(t, sv.Equal(nil))
	assert.True(t, sv.Equal(sv))
	assert.False(t, sv.Equal(MustSemanticVersion("v0.1.0")))

	assert.False(t, sv.Less(nil))
	assert.False(t, sv.Less(MustSemanticVersion("v0.1.0")))
	assert.True(t, sv.Less(MustSemanticVersion("v2.0.0")))

	_, err = NewSemanticVersion("xxx")
	require.Error(t, err)

	assert.Panics(t, func() {
		MustSemanticVersion("xxx")
	})
}

func TestSequenceVersion(t *testing.T) {
	sv, err := NewSequenceVersion("123")
	require.NoError(t, err)

	assert.False(t, sv.Equal(nil))
	assert.True(t, sv.Equal(sv))
	assert.False(t, sv.Equal(MustSequenceVersion("120")))

	assert.False(t, sv.Less(nil))
	assert.False(t, sv.Less(MustSequenceVersion("145")))
	assert.True(t, sv.Less(MustSequenceVersion("14")))

	_, err = NewSequenceVersion("xxx")
	require.Error(t, err)

	assert.Panics(t, func() {
		MustSequenceVersion("xxx")
	})

	assert.Equal(t, "123", sv.String())
}
