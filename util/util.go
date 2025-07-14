package util

import "os"

func IsEnvProd() bool {
	return os.Getenv("ENV") == "production"
}
