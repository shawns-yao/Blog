package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadRootEnv loads the project-root environment files without replacing process overrides.
func LoadRootEnv() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if info, statErr := os.Stat(filepath.Join(dir, "server", "go.mod")); statErr == nil && !info.IsDir() {
			var files []string
			for _, name := range []string{".env", ".env.rag"} {
				path := filepath.Join(dir, name)
				if _, statErr := os.Stat(path); statErr == nil {
					files = append(files, path)
				} else if !os.IsNotExist(statErr) {
					return statErr
				}
			}
			if len(files) == 0 {
				return nil
			}
			return godotenv.Load(files...)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}
