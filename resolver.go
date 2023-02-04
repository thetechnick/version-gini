package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-air/gini"
	"github.com/go-air/gini/z"
)

// Records a resolver run.
type Resolver struct {
	db ProjectDB

	resolveOnce               sync.Once
	resolved                  []ResolverProjectVersion
	gini                      *gini.Gini
	projectConstraints        map[string][]ResolverConstraint
	projectVersionsToLiterals map[ResolverProjectVersion]z.Lit
}

type ResolverProjectVersion struct {
	Name    string
	Version Version
}

func (rpv ResolverProjectVersion) String() string {
	return rpv.Name + "=" + rpv.Version.String()
}

// Records constraints for the resolver and their source.
type ResolverConstraint struct {
	// Project and Version that is the source of these constraints.
	Origin ResolverProjectVersion
	// ProjectName the constrain targets.
	SubjectProjectName string
	Constraints        []Constraint
}

func (rc ResolverConstraint) String() string {
	var constraints []string
	for _, c := range rc.Constraints {
		constraints = append(constraints, c.String())
	}

	return fmt.Sprintf(
		"%s constrains %q with %s",
		rc.Origin, rc.SubjectProjectName, strings.Join(constraints, ", "))
}

func NewResolver(db ProjectDB) *Resolver {
	return &Resolver{
		db: db,

		gini:                      gini.New(),
		projectConstraints:        map[string][]ResolverConstraint{},
		projectVersionsToLiterals: map[ResolverProjectVersion]z.Lit{},
	}
}

func (r *Resolver) Resolve(ctx context.Context, rootDeps []Dependency) ([]ResolverProjectVersion, error) {
	var err error
	r.resolveOnce.Do(func() {
		err = r.resolve(ctx, rootDeps)
	})
	return r.resolved, err
}

func (r *Resolver) ConstrainsFor(ctx context.Context, projectName string) []ResolverConstraint {
	return r.projectConstraints[projectName]
}

func (r *Resolver) resolve(ctx context.Context, rootDeps []Dependency) error {
	// 1.
	// Discover projects and constraints that are part of the dependency tree.
	if err := r.walkProjectConstraints(ctx,
		Project{
			Name: "root",
			Versions: []ProjectVersion{
				{Dependencies: rootDeps},
			},
		}); err != nil {
		return err
	}

	// 2.
	// Assign each project version a literal for the SAT solver.
	var i int
	for projectName := range r.projectConstraints {
		project, err := r.db.Get(ctx, projectName)
		if err != nil {
			return err
		}
		for _, pv := range project.Versions {
			i++
			lit := z.Var(i).Pos()
			r.projectVersionsToLiterals[ResolverProjectVersion{
				Name:    projectName,
				Version: pv.Version,
			}] = lit
			// build first constraint clause
			// we want at least one version of each project
			r.gini.Add(lit)
		}
		r.gini.Add(0)
	}

	// 3.
	// Apply dependency constraints
	for projectName, constraints := range r.projectConstraints {
		project, err := r.db.Get(ctx, projectName)
		if err != nil {
			return err
		}
		for _, constraint := range constraints {
			srcLit := r.projectVersionsToLiterals[constraint.Origin]
			for _, pv := range project.Versions {
				if ConstraintAND(constraint.Constraints).Matches(pv.Version) {
					// matches -> unconstrained!
					continue
				}
				r.gini.Add(srcLit.Not())
				r.gini.Add(r.projectVersionsToLiterals[ResolverProjectVersion{
					Name:    projectName,
					Version: pv.Version,
				}].Not())
				r.gini.Add(0)
			}
		}
	}

	if r.gini.Solve() != 1 {
		return fmt.Errorf("nosat")
	}

	for pv, lit := range r.projectVersionsToLiterals {
		if r.gini.Value(lit) {
			r.resolved = append(r.resolved, pv)
		}
	}

	return nil
}

func (r *Resolver) walkProjectConstraints(
	ctx context.Context,
	project Project,
) error {
	for _, pv := range project.Versions {
		for _, dep := range pv.Dependencies {
			if _, ok := r.projectConstraints[dep.Name]; !ok {
				// ensure key is present, even if unconstrained.
				r.projectConstraints[dep.Name] = nil
			}

			if len(dep.Constraints) != 0 {
				r.projectConstraints[dep.Name] = append(
					r.projectConstraints[dep.Name],
					ResolverConstraint{
						Origin: ResolverProjectVersion{
							Name:    project.Name,
							Version: pv.Version,
						},
						SubjectProjectName: dep.Name,
						Constraints:        dep.Constraints,
					},
				)
			}

			depProject, err := r.db.Get(ctx, dep.Name)
			if err != nil {
				return err
			}
			if err := r.walkProjectConstraints(
				ctx, depProject); err != nil {
				return err
			}
		}
	}
	return nil
}

type ResolverProjectVersionByName []ResolverProjectVersion

func (a ResolverProjectVersionByName) Len() int           { return len(a) }
func (a ResolverProjectVersionByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ResolverProjectVersionByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
