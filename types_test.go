package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
