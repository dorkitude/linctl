package cmd

import (
	"testing"
)

func TestInitiativeCommandRegistration(t *testing.T) {
	// Verify the initiative command is registered on the root command
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "initiative" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("initiative command not registered on root")
	}
}

func TestInitiativeSubcommands(t *testing.T) {
	expectedSubcommands := []string{
		"list",
		"get",
		"create",
		"update",
		"delete",
		"archive",
		"unarchive",
		"link",
		"unlink",
	}

	subcommandNames := make(map[string]bool)
	for _, cmd := range initiativeCmd.Commands() {
		subcommandNames[cmd.Name()] = true
	}

	for _, expected := range expectedSubcommands {
		if !subcommandNames[expected] {
			t.Errorf("missing subcommand %q on initiative command", expected)
		}
	}
}

func TestInitiativeListAliases(t *testing.T) {
	found := false
	for _, alias := range initiativeListCmd.Aliases {
		if alias == "ls" {
			found = true
			break
		}
	}
	if !found {
		t.Error("initiative list command should have 'ls' alias")
	}
}

func TestInitiativeGetAliases(t *testing.T) {
	found := false
	for _, alias := range initiativeGetCmd.Aliases {
		if alias == "show" {
			found = true
			break
		}
	}
	if !found {
		t.Error("initiative get command should have 'show' alias")
	}
}

func TestInitiativeCreateFlags(t *testing.T) {
	requiredFlags := []string{"name"}
	for _, name := range requiredFlags {
		flag := initiativeCreateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing required flag %q on initiative create", name)
		}
	}

	optionalFlags := []string{"description", "status", "owner", "target-date", "color"}
	for _, name := range optionalFlags {
		flag := initiativeCreateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("missing optional flag %q on initiative create", name)
		}
	}
}

func TestInitiativeUpdateFlags(t *testing.T) {
	flags := []string{"name", "description", "status", "owner", "target-date", "color"}
	for _, name := range flags {
		flag := initiativeUpdateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("missing flag %q on initiative update", name)
		}
	}
}

func TestInitiativeDeleteFlags(t *testing.T) {
	flag := initiativeDeleteCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("missing flag 'force' on initiative delete")
	}
}

func TestInitiativeLinkFlags(t *testing.T) {
	flag := initiativeLinkCmd.Flags().Lookup("project")
	if flag == nil {
		t.Fatal("missing flag 'project' on initiative link")
	}
}

func TestInitiativeUnlinkFlags(t *testing.T) {
	flag := initiativeUnlinkCmd.Flags().Lookup("project")
	if flag == nil {
		t.Fatal("missing flag 'project' on initiative unlink")
	}
}

func TestNormalizeInitiativeStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"planned", "Planned"},
		{"Planned", "Planned"},
		{"PLANNED", "Planned"},
		{"active", "Active"},
		{"Active", "Active"},
		{"completed", "Completed"},
		{"invalid", ""},
		{"", ""},
	}

	for _, tc := range tests {
		result := normalizeInitiativeStatus(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeInitiativeStatus(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}
