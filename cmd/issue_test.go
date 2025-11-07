package cmd

import (
    "errors"
    "testing"
)

func TestIsValidUUID(t *testing.T) {
    valid := []string{
        "123e4567-e89b-12d3-a456-426614174000",
        "00000000-0000-0000-0000-000000000000",
        "ABCDEFAB-CDEF-ABCD-EFAB-CDEFABCDEFAB",
    }
    invalid := []string{
        "", "unassigned", "1234", "g23e4567-e89b-12d3-a456-426614174000",
        "123e4567e89b12d3a456426614174000",
        "123e4567-e89b-12d3-a456-426614174000-extra",
    }

    for _, v := range valid {
        if !isValidUUID(v) {
            t.Errorf("expected valid UUID: %s", v)
        }
    }
    for _, v := range invalid {
        if isValidUUID(v) {
            t.Errorf("expected invalid UUID: %s", v)
        }
    }
}

func TestBuildProjectInput(t *testing.T) {
    // Empty → ok=false, no input
    if val, ok, err := buildProjectInput(""); err != nil || ok || val != nil {
        t.Errorf("empty flag: want (nil,false,nil), got (%v,%v,%v)", val, ok, err)
    }

    // unassigned → ok=true, val=nil
    if val, ok, err := buildProjectInput("unassigned"); err != nil || !ok || val != nil {
        t.Errorf("unassigned: want (nil,true,nil), got (%v,%v,%v)", val, ok, err)
    }

    // valid uuid → ok=true, val=uuid
    uuid := "123e4567-e89b-12d3-a456-426614174000"
    if val, ok, err := buildProjectInput(uuid); err != nil || !ok || val != uuid {
        t.Errorf("uuid: want (%s,true,nil), got (%v,%v,%v)", uuid, val, ok, err)
    }

    // invalid uuid → error
    if _, _, err := buildProjectInput("not-a-uuid"); err == nil {
        t.Errorf("expected error for invalid uuid")
    }
}

func TestIsProjectNotFoundErr(t *testing.T) {
    cases := []struct {
        in   error
        want bool
    }{
        {errors.New("GraphQL errors: [{ message: 'Project not found' }]"), true},
        {errors.New("something about projectId not found"), true},
        {errors.New("issue not found"), false},
        {errors.New("unknown error"), false},
        {nil, false},
    }
    for _, c := range cases {
        got := isProjectNotFoundErr(c.in)
        if got != c.want {
            t.Errorf("isProjectNotFoundErr(%v) = %v, want %v", c.in, got, c.want)
        }
    }
}

