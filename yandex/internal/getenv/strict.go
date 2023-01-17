package getenv

import (
	"fmt"
	"os"
)

func Strict(envName string) string {
	value := os.Getenv(envName)

	if value == "" {
		panic(fmt.Errorf("%s environment variable is not defiend", envName))
	}

	return value
}
