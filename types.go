package main

type Project struct {
	Name     string
	Versions []ProjectVersion
}

type ProjectVersion struct {
	Version      Version
	Dependencies []Dependency
}

type ProjectVersionsDescending []ProjectVersion

func (a ProjectVersionsDescending) Len() int           { return len(a) }
func (a ProjectVersionsDescending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ProjectVersionsDescending) Less(i, j int) bool { return a[j].Version.Less(a[i].Version) }

type Dependency struct {
	Name        string
	Constraints []Constraint
}

// ConstraintAND is AND of all constraints.
type ConstraintAND []Constraint

func (c ConstraintAND) Matches(v Version) bool {
	for _, con := range c {
		if !con.Matches(v) {
			return false
		}
	}
	return true
}
