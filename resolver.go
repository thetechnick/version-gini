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
	projects                  []Project
	projectConstraints        map[string][]ResolverConstraint
	projectVersionsToLiterals map[ResolverProjectVersion]z.Lit
}

type ResolverProjectVersion struct {
	Name    string
	Version string
}

func (rpv ResolverProjectVersion) String() string {
	if len(rpv.Version) == 0 {
		return rpv.Name
	}
	return rpv.Name + "=" + rpv.Version
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
		err = r.setup(ctx, rootDeps)
		if err != nil {
			return
		}
		err = r.resolve(ctx, rootDeps)
	})
	return r.resolved, err
}

func (r *Resolver) ConstrainsFor(ctx context.Context, projectName string) []ResolverConstraint {
	return r.projectConstraints[projectName]
}

func (r *Resolver) setup(ctx context.Context, rootDeps []Dependency) error {
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
	for _, project := range r.projects {
		// CONSTRAINT: We want at least one version of each project
		for _, pv := range project.Versions {
			lit := r.gini.Lit()
			r.projectVersionsToLiterals[ResolverProjectVersion{
				Name:    project.Name,
				Version: pv.Version.String(),
			}] = lit
			r.gini.Add(lit)
		}
		r.gini.Add(z.LitNull)

		// CONSTRAINT: We want at MOST one version of each project
		for _, pv := range project.Versions {
			rPV := ResolverProjectVersion{
				Name:    project.Name,
				Version: pv.Version.String(),
			}
			pvLit := r.projectVersionsToLiterals[rPV]
			for _, otherPV := range project.Versions {
				otherRPV := ResolverProjectVersion{
					Name:    project.Name,
					Version: otherPV.Version.String(),
				}
				if rPV == otherRPV {
					// We don't want to exclude ourselves!
					continue
				}
				otherPVLit := r.projectVersionsToLiterals[otherRPV]
				r.gini.Add(pvLit.Not())
				r.gini.Add(otherPVLit.Not())
				r.gini.Add(z.LitNull)
			}
		}

		// CONSTRAINT: Process actual dependency constraints
		constraints := r.projectConstraints[project.Name]
		for _, constraint := range constraints {
			for _, pv := range project.Versions {
				if ConstraintAND(constraint.Constraints).Matches(pv.Version) {
					// matches -> unconstrained!
					continue
				}
				srcLit := r.projectVersionsToLiterals[constraint.Origin]
				if srcLit != 0 {
					r.gini.Add(srcLit.Not())
				}
				r.gini.Add(r.projectVersionsToLiterals[ResolverProjectVersion{
					Name:    constraint.SubjectProjectName,
					Version: pv.Version.String(),
				}].Not())
				r.gini.Add(z.LitNull)
			}
		}
	}

	return nil
}

func (r *Resolver) resolve(ctx context.Context, rootDeps []Dependency) error {
	// Shortcut, is there any combination that works?
	if r.gini.Solve() != 1 {
		return fmt.Errorf("nosat!")
	}

	// selectedProjectVersion := map[string]Version{}
	selectedProjectVersion := map[string]Version{}
	var (
		projectIndex        int
		projectVersionIndex int
	)

	// Find _latest_ version of all components that still satisfy the model, by
	// starting with the latest version of each project and testing older and older versions.
tryAgain:
	// select version to try:
	if projectIndex >= len(r.projects) {
		return fmt.Errorf("NOSAT! out of projects!")
	}
	project := r.projects[projectIndex]
	if _, ok := selectedProjectVersion[project.Name]; !ok {
		if projectVersionIndex >= len(project.Versions) {
			return fmt.Errorf("NOSAT! out of versions for %s", project.Name)
		}

		selectedProjectVersion[project.Name] = project.Versions[projectVersionIndex].Version
	}

	for projectName, version := range selectedProjectVersion {
		r.gini.Assume(r.projectVersionsToLiterals[ResolverProjectVersion{
			Name:    projectName,
			Version: version.String(),
		}])
	}

	if r.gini.Solve() != 1 {
		// select next version when UNSAT
		delete(selectedProjectVersion, project.Name)
		projectVersionIndex++
		goto tryAgain
	}

	// do we have a solution for all projects?
	if len(selectedProjectVersion) != len(r.projects) {
		// add next project
		projectIndex++
		projectVersionIndex = 0
		goto tryAgain
	}

	var resolved []ResolverProjectVersion
	for pv, lit := range r.projectVersionsToLiterals {
		if r.gini.Value(lit) {
			resolved = append(resolved, pv)
		}
	}
	r.resolved = resolved
	return nil
}

func (r *Resolver) walkProjectConstraints(
	ctx context.Context,
	project Project,
) error {

	for _, pv := range project.Versions {
		for _, dep := range pv.Dependencies {
			depProject, err := r.db.Get(ctx, dep.Name)
			if err != nil {
				return err
			}

			if _, ok := r.projectConstraints[dep.Name]; !ok {
				// ensure key is present, even if unconstrained.
				r.projectConstraints[dep.Name] = nil
				r.projects = append(r.projects, depProject)
			}

			if len(dep.Constraints) != 0 {
				var v string
				if pv.Version != nil {
					v = pv.Version.String()
				}
				r.projectConstraints[dep.Name] = append(
					r.projectConstraints[dep.Name],
					ResolverConstraint{
						Origin: ResolverProjectVersion{
							Name:    project.Name,
							Version: v,
						},
						SubjectProjectName: dep.Name,
						Constraints:        dep.Constraints,
					},
				)
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
