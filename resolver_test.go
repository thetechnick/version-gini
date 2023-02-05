package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver(t *testing.T) {
	t.SkipNow()
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
		{Name: "A", Version: "1.1.0"},
		{Name: "B", Version: "1.0.0"},
		{Name: "C", Version: "2.0.0"},
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

// generateProjectDBEntries generates a number of projects and project versions for testing and benchmarks.
// every project version has dependencies on all other projects of the same version.
func generateProjectDBEntries(projectNumber, projectVersions int) []Project {
	var projects []Project
	for i := 0; i < projectNumber; i++ {
		project := Project{
			Name: fmt.Sprintf("P%d", i),
		}
		v := MustSemanticVersion("v1.0.0")
		for j := 0; j < projectVersions; j++ {
			pv := ProjectVersion{
				Version: v,
			}

			if i == 0 {
				for di := 0; di < projectNumber; di++ {
					if di == 0 {
						continue
					} // skip dependency on yourself

					pv.Dependencies = append(pv.Dependencies, Dependency{
						Name:        fmt.Sprintf("P%d", di),
						Constraints: []Constraint{*NewConstraint(Equal, v)},
					})
				}
			}
			project.Versions = append(project.Versions, pv)

			nextVersion := v.IncMinor()
			v = &SemanticVersion{Version: &nextVersion}
		}

		projects = append(projects, project)
	}
	return projects
}

func TestXxx(t *testing.T) {
	projects := generateProjectDBEntries(1, 3)

	ctx := context.Background()
	db := NewInMemoryDB()
	for _, p := range projects {
		require.NoError(t, db.Add(ctx, p))
	}

	j, _ := json.Marshal(projects)
	fmt.Printf("%s\n", j)

	r := NewResolver(db)
	result, err := r.Resolve(ctx, []Dependency{
		{Name: "P0", Constraints: []Constraint{
			// *NewConstraint(Equal, MustSemanticVersion("v1.0.0")),
		}},
	})
	require.NoError(t, err)

	sort.Sort(ResolverProjectVersionByName(result))
	fmt.Println(result)

	t.Fail()
}
