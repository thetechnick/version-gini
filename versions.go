package main

import (
	"strconv"

	"github.com/Masterminds/semver/v3"
)

type SemanticVersion struct {
	*semver.Version
}

var (
	_ Version = (*SemanticVersion)(nil)
)

func MustSemanticVersion(v string) Version {
	sv, err := NewSemanticVersion(v)
	if err != nil {
		panic(err)
	}
	return sv
}

func NewSemanticVersion(v string) (Version, error) {
	sv, err := semver.NewVersion(v)
	if err != nil {
		return nil, err
	}

	return &SemanticVersion{Version: sv}, nil
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

type SequenceVersion int

var (
	_ Version = SequenceVersion(0)
)

func MustSequenceVersion(v string) Version {
	sv, err := NewSequenceVersion(v)
	if err != nil {
		panic(err)
	}
	return sv
}

func NewSequenceVersion(v string) (Version, error) {
	s, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}

	return SequenceVersion(s), nil
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
