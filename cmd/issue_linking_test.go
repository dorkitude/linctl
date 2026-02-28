package cmd

import (
	"testing"

	"github.com/dorkitude/linctl/pkg/api"
)

func TestIsUnsetValue(t *testing.T) {
	cases := map[string]bool{
		"":           true,
		"none":       true,
		"None":       true,
		"null":       true,
		"unassigned": true,
		"project-a":  false,
	}

	for input, want := range cases {
		if got := isUnsetValue(input); got != want {
			t.Fatalf("isUnsetValue(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestFindProjectByNameOrID(t *testing.T) {
	projects := []api.Project{
		{ID: "proj_1", Name: "Q1 Platform"},
		{ID: "proj_2", Name: "Revenue"},
	}

	byID := findProjectByNameOrID(projects, "proj_2")
	if byID == nil || byID.ID != "proj_2" {
		t.Fatalf("expected to resolve project by ID")
	}

	byName := findProjectByNameOrID(projects, "q1 platform")
	if byName == nil || byName.ID != "proj_1" {
		t.Fatalf("expected to resolve project by case-insensitive name")
	}

	none := findProjectByNameOrID(projects, "missing")
	if none != nil {
		t.Fatalf("expected nil for unknown project")
	}
}

func TestFindMilestoneByNameOrID(t *testing.T) {
	milestones := []api.ProjectMilestone{
		{ID: "mil_1", Name: "Phase 1"},
		{ID: "mil_2", Name: "GA"},
	}

	byID := findMilestoneByNameOrID(milestones, "mil_2")
	if byID == nil || byID.ID != "mil_2" {
		t.Fatalf("expected to resolve milestone by ID")
	}

	byName := findMilestoneByNameOrID(milestones, "phase 1")
	if byName == nil || byName.ID != "mil_1" {
		t.Fatalf("expected to resolve milestone by case-insensitive name")
	}

	none := findMilestoneByNameOrID(milestones, "missing")
	if none != nil {
		t.Fatalf("expected nil for unknown milestone")
	}
}
