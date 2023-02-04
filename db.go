package main

import (
	"context"
	"errors"
	"sort"
)

type ProjectDB interface {
	Add(ctx context.Context, project Project) error
	Get(ctx context.Context, projectName string) (Project, error)
}

var (
	ErrNotFound = errors.New("not found")
)

type InMemoryDB struct {
	// data indexed by project name
	data map[string]Project
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		data: map[string]Project{},
	}
}

func (db *InMemoryDB) Add(ctx context.Context, project Project) error {
	sort.Sort(ProjectVersionsDescending(project.Versions))
	db.data[project.Name] = project
	return nil
}

func (db *InMemoryDB) Get(ctx context.Context, projectName string) (Project, error) {
	project, ok := db.data[projectName]
	if !ok {
		return Project{}, ErrNotFound
	}
	return project, nil
}
