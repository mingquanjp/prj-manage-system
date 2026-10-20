package config

import "testing"

func TestRequireDev(t *testing.T) {
	for _, env := range []string{"", "prod", "test", "DEV", "dev"} {
		t.Run("env="+env, func(t *testing.T) {
			t.Setenv("APP_ENV", env)
			err := RequireDev("thao tác mẫu")
			if (err == nil) != (env == "dev") {
				t.Fatalf("RequireDev với APP_ENV=%q: %v", env, err)
			}
		})
	}
}
