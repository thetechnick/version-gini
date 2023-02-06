package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstraint(t *testing.T) {
	tests := []struct {
		op          Operator
		versionBase Version
		versionTest Version
		result      bool
	}{
		{
			op:          Equal,
			versionBase: MustSemanticVersion("v1.0.0"),
			versionTest: MustSemanticVersion("v1.0.0"),
			result:      true,
		},
		{
			op:          NotEqual,
			versionBase: MustSemanticVersion("v1.0.0"),
			versionTest: MustSemanticVersion("v1.0.0"),
			result:      false,
		},
		{
			op:          Greater,
			versionBase: MustSemanticVersion("v1.6.0"),
			versionTest: MustSemanticVersion("v1.2.0"),
			result:      true,
		},
		{
			op:          Less,
			versionBase: MustSemanticVersion("v1.0.0"),
			versionTest: MustSemanticVersion("v1.2.0"),
			result:      true,
		},
		{
			op:          GreaterOrEqual,
			versionBase: MustSemanticVersion("v1.5.0"),
			versionTest: MustSemanticVersion("v1.2.0"),
			result:      true,
		},
		{
			op:          LessOrEqual,
			versionBase: MustSemanticVersion("v1.1.0"),
			versionTest: MustSemanticVersion("v1.2.0"),
			result:      true,
		},
	}

	for _, test := range tests {
		t.Run(string(test.op), func(t *testing.T) {
			c := NewConstraint(test.op, test.versionBase)
			assert.Equal(t,
				test.result, c.Matches(test.versionTest))
		})
	}
}

func TestParseConstraint(t *testing.T) {
	c, err := ParseConstraint(">v1.2.3", ParseSemanticVersion)
	require.NoError(t, err)

	assert.Equal(t, Greater, c.operator)
	assert.Equal(t, "1.2.3", c.version.String())
}

func TestParseConstraint_unknownOp(t *testing.T) {
	_, err := ParseConstraint("()v1.2.3", ParseSemanticVersion)
	require.EqualError(t, err, "unknown op")
}

func TestParseConstraint_parserError(t *testing.T) {
	_, err := ParseConstraint("=vxxx", ParseSemanticVersion)
	require.EqualError(t, err, "Invalid Semantic Version")
}
