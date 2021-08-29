package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	KubeConfigPath string
)

func init() {
	KubeConfigPath = os.Getenv("KUBE_CONFIG_PATH")
}
