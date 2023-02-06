package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Constraint struct {
	operator Operator
	version  Version
}

func NewConstraint(op Operator, v Version) *Constraint {
	return &Constraint{operator: op, version: v}
}

type VersionParser func(v string) (Version, error)

func ParseConstraint(constraint string, parseVersion VersionParser) (*Constraint, error) {
	for _, op := range []Operator{Equal, NotEqual, Greater, Less, GreaterOrEqual, LessOrEqual} {
		if strings.HasPrefix(constraint, string(op)) {
			v, err := parseVersion(constraint[len(op):])
			if err != nil {
				return nil, err
			}
			return NewConstraint(op, v), nil
		}
	}
	return nil, fmt.Errorf("unknown op")
}

func (c *Constraint) String() string {
	return string(c.operator) + c.version.String()
}

func (c *Constraint) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
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
