package test_utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
)

var loadOnce sync.Once

func LoadTestEnv(envFileName ...string) {
	loadOnce.Do(func() {
		fileName := ".env.test"
		if len(envFileName) > 0 {
			fileName = envFileName[0]
		}

		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
		} else {
			for {
				envFile := filepath.Join(dir, fileName)
				if _, err := os.Stat(envFile); err == nil {
					if err := godotenv.Load(envFile); err != nil {
						fmt.Printf("Error loading %s file: %v\n", fileName, err)
					} else {
						fmt.Printf("Loaded %s from: %s\n", fileName, envFile)
						break
					}
				}

				parent := filepath.Dir(dir)
				if parent == dir {
					fmt.Printf("Reached root directory, %s not found\n", fileName)
					break
				}
				dir = parent
			}
		}
	})
}
