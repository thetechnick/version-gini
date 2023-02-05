package main

import "encoding/json"

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
func (a ProjectVersionsDescending) Less(i, j int) bool { return a[i].Version.Less(a[j].Version) }

type Dependency struct {
	Name        string
	Constraints []Constraint
}

type Version interface {
	Equal(v Version) bool
	Less(v Version) bool
	String() string
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

type Constraint struct {
	operator Operator
	version  Version
}

func (c *Constraint) String() string {
	return string(c.operator) + c.version.String()
}

func (c *Constraint) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func NewConstraint(op Operator, v Version) *Constraint {
	return &Constraint{operator: op, version: v}
}

func (c *Constraint) Matches(v Version) bool {
	switch c.operator {
	case Equal:
		return c.version.Equal(v)
	case NotEqual:
		return !c.version.Equal(v)
	case Greater:
		return !c.version.Less(v) && !c.version.Equal(v)
	case Less:
		return c.version.Less(v)
	case GreaterOrEqual:
		return !c.version.Less(v)
	case LessOrEqual:
		return c.version.Less(v) || c.version.Equal(v)
	default:
		return false
	}
}

type Operator string

const (
	Equal          Operator = "="
	NotEqual       Operator = "!="
	Greater        Operator = ">"
	Less           Operator = "<"
	GreaterOrEqual Operator = ">="
	LessOrEqual    Operator = "<="
)
