package main

import (
	"strconv"

	"github.com/Masterminds/semver/v3"
)

type Version interface {
	Equal(v Version) bool
	Less(v Version) bool
	String() string
}

// Represents a Semantic Version v2.
type SemanticVersion struct {
	*semver.Version
}

var (
	_ Version = (*SemanticVersion)(nil)
)

func MustSemanticVersion(v string) *SemanticVersion {
	sv, err := NewSemanticVersion(v)
	if err != nil {
		panic(err)
	}
	return sv
}

func NewSemanticVersion(v string) (*SemanticVersion, error) {
	sv, err := semver.NewVersion(v)
	if err != nil {
		return nil, err
	}

	return &SemanticVersion{Version: sv}, nil
}

func ParseSemanticVersion(v string) (Version, error) {
	return NewSemanticVersion(v)
}

func (sv *SemanticVersion) Equal(v Version) bool {
	otherSV, ok := v.(*SemanticVersion)
	if !ok {
		return false
	}
	return sv.Version.Equal(otherSV.Version)
}

func (sv *SemanticVersion) Less(v Version) bool {
	otherSV, ok := v.(*SemanticVersion)
	if !ok {
		return false
	}
	return sv.Version.LessThan(otherSV.Version)
}

// Sequence version is just an increasing number.
type SequenceVersion int

var (
	_ Version = SequenceVersion(0)
)

func MustSequenceVersion(v string) SequenceVersion {
	sv, err := NewSequenceVersion(v)
	if err != nil {
		panic(err)
	}
	return sv
}

func NewSequenceVersion(v string) (SequenceVersion, error) {
	s, err := strconv.Atoi(v)
	if err != nil {
		return SequenceVersion(0), err
	}

	return SequenceVersion(s), nil
}

func ParseSequenceVersion(v string) (Version, error) {
	return NewSequenceVersion(v)
}

func (sv SequenceVersion) Equal(v Version) bool {
	otherSV, ok := v.(SequenceVersion)
	if !ok {
		return false
	}
	return int(otherSV) == int(sv)
}

func (sv SequenceVersion) Less(v Version) bool {
	otherSV, ok := v.(SequenceVersion)
	if !ok {
		return false
	}
	return int(otherSV) < int(sv)
}

func (sv SequenceVersion) String() string {
	return strconv.Itoa(int(sv))
}
