package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	required := []string{"APP_ENV", "APP_ADDR"}
	optionalSecrets := []string{"MYSQL_DSN", "REDIS_PASSWORD", "JWT_SECRET"}
	failed := false

	fmt.Printf("go=%s os=%s arch=%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	for _, key := range required {
		if os.Getenv(key) == "" {
			fmt.Printf("MISSING required variable: %s\n", key)
			failed = true
			continue
		}
		fmt.Printf("OK required variable: %s\n", key)
	}
	for _, key := range optionalSecrets {
		if os.Getenv(key) == "" {
			fmt.Printf("NOT SET optional secret: %s\n", key)
			continue
		}
		fmt.Printf("SET optional secret: %s (value hidden)\n", key)
	}

	if failed {
		os.Exit(1)
	}
}
