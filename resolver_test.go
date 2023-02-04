package main

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver(t *testing.T) {
	projectA := Project{
		Name: "A",
		Versions: []ProjectVersion{
			{
				Version: MustSemanticVersion("1.1.1"),
				Dependencies: []Dependency{
					{Name: "C", Constraints: []Constraint{
						*NewConstraint(Equal, MustSemanticVersion("2.0.1")),
					}},
				},
			},
			{
				Version: MustSemanticVersion("1.1.0"),
				Dependencies: []Dependency{
					{Name: "C", Constraints: []Constraint{
						*NewConstraint(Equal, MustSemanticVersion("2.0.0")),
					}},
				},
			},
		},
	}
	projectB := Project{
		Name: "B",
		Versions: []ProjectVersion{
			{
				Version: MustSemanticVersion("1.0.0"),
				Dependencies: []Dependency{
					{Name: "C", Constraints: []Constraint{
						*NewConstraint(Equal, MustSemanticVersion("2.0.0")),
					}},
				},
			},
		},
	}
	projectC := Project{
		Name: "C",
		Versions: []ProjectVersion{
			{
				Version: MustSemanticVersion("2.0.1"),
			},
			{
				Version: MustSemanticVersion("2.0.0"),
			},
		},
	}

	db := NewInMemoryDB()
	ctx := context.Background()

	require.NoError(t, db.Add(ctx, projectA))
	require.NoError(t, db.Add(ctx, projectB))
	require.NoError(t, db.Add(ctx, projectC))

	r := NewResolver(db)
	result, err := r.Resolve(ctx, []Dependency{
		{Name: "A"},
		{Name: "B"},
	})
	require.NoError(t, err)

	// Assertions
	sort.Sort(ResolverProjectVersionByName(result))
	t.Log(result)

	assert.Equal(t, []ResolverProjectVersion{
		{Name: "A", Version: MustSemanticVersion("1.1.0")},
		{Name: "B", Version: MustSemanticVersion("1.0.0")},
		{Name: "C", Version: MustSemanticVersion("2.0.0")},
	}, result)

	constraints := r.ConstrainsFor(ctx, "C")
	var constraintStrings []string
	for _, c := range constraints {
		constraintStrings = append(constraintStrings, c.String())
	}
	t.Log("C is constrained by:\n" + strings.Join(constraintStrings, "\n"))

	assert.Contains(t, constraintStrings, `A=1.1.0 constrains "C" with =2.0.0`)
	assert.Contains(t, constraintStrings, `A=1.1.1 constrains "C" with =2.0.1`)
	assert.Contains(t, constraintStrings, `B=1.0.0 constrains "C" with =2.0.0`)
}
