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
