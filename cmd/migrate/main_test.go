package main

import (
	"os"
	"strings"
	"testing"
)

func TestValidateCommand(t *testing.T) {
	for _, env := range []string{"", "prod", "dev"} {
		for _, command := range []string{"up", "version", "down", "down-all", "drop", "invalid"} {
			t.Run(env+"/"+command, func(t *testing.T) {
				t.Setenv("APP_ENV", env)
				allowed := command == "up" || command == "version" ||
					(env == "dev" && (command == "down" || command == "down-all"))
				if err := validateCommand(command); (err == nil) != allowed {
					t.Fatalf("validateCommand(%q): %v", command, err)
				}
			})
		}
	}
}

func TestRunRejectsDestructiveCommandBeforeConfig(t *testing.T) {
	previous := os.Args
	t.Cleanup(func() { os.Args = previous })
	t.Setenv("APP_ENV", "prod")
	t.Setenv("DATABASE_URL", "")
	for _, command := range []string{"down", "down-all"} {
		os.Args = []string{"migrate", command}
		if err := run(); err == nil || !strings.Contains(err.Error(), "APP_ENV=dev") {
			t.Fatalf("phải từ chối trước khi đọc cấu hình DB: %v", err)
		}
	}
}

func TestMigrationArguments(t *testing.T) {
	for _, test := range []struct {
		name    string
		args    []string
		command string
		actor   int
		valid   bool
	}{
		{"default", nil, "up", 0, true},
		{"up", []string{"up"}, "up", 0, true},
		{"explicit import actor", []string{"up", "--msk-import-actor=101"}, "up", 101, true},
		{"max user id", []string{"up", "--msk-import-actor", "2147483647"}, "up", 2147483647, true},
		{"down without actor", []string{"down"}, "down", 0, true},
		{"negative actor", []string{"up", "--msk-import-actor=-1"}, "", 0, false},
		{"overflow actor", []string{"up", "--msk-import-actor=2147483648"}, "", 0, false},
		{"actor is not a number", []string{"up", "--msk-import-actor=employee"}, "", 0, false},
		{"missing actor", []string{"up", "--msk-import-actor"}, "", 0, false},
		{"unknown option", []string{"up", "--force"}, "", 0, false},
		{"unexpected argument", []string{"up", "101"}, "", 0, false},
		{"down actor", []string{"down", "--msk-import-actor=101"}, "", 0, false},
		{"version actor", []string{"version", "--msk-import-actor=101"}, "", 0, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			command, actor, err := migrationArguments(test.args)
			if (err == nil) != test.valid || command != test.command || actor != test.actor {
				t.Fatalf("đối số migration không đúng contract: command=%q actor=%d err=%v", command, actor, err)
			}
		})
	}
}

func TestRunRejectsInvalidImportActorBeforeConfig(t *testing.T) {
	previous := os.Args
	t.Cleanup(func() { os.Args = previous })
	t.Setenv("DATABASE_URL", "")
	os.Args = []string{"migrate", "up", "--msk-import-actor=-1"}
	if err := run(); err == nil || !strings.Contains(err.Error(), "actor import") {
		t.Fatal("phải từ chối actor không hợp lệ trước khi đọc cấu hình DB")
	}
}
