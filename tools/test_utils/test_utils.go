package test_utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
)

var loadOnce sync.Once

func LoadTestEnv() {
	loadOnce.Do(func() {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current working directory: %v\n", err)
		} else {
			for {
				envFile := filepath.Join(dir, ".env.test")
				if _, err := os.Stat(envFile); err == nil {
					if err := godotenv.Load(envFile); err != nil {
						fmt.Printf("Error loading .env.test file: %v\n", err)
					} else {
						fmt.Printf("Loaded .env.test from: %s\n", envFile)
						break
					}
				}

				parent := filepath.Dir(dir)
				if parent == dir {
					fmt.Println("Reached root directory, .env.test not found")
					break
				}
				dir = parent
			}
		}
	})
}
