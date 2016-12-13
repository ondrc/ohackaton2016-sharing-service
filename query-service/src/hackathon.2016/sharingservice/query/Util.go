package main

import (
	"log"
	"os"
)

func MustGetEnv(name string) string {
	env := os.Getenv(name)
	if env == "" {
		log.Fatalf("Required environment variable is not defined or empty: %v \n", name)
	}
	return env
}

func GetEnvOr(name, defValue string) string {
	env := os.Getenv(name)
	if env == "" {
		env = defValue
	}
	return env
}