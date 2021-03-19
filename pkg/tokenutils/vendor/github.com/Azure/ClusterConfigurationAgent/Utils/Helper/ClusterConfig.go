package Helper

import (
	"os"
	"os/user"
	"path/filepath"
)

var (
	DefaultKubeConfig = filepath.Join(".kube","config")
)

func GetLocalClusterPath() string {
	file := os.Getenv("KUBECONFIG")
	if file == "" {
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		return filepath.Join(usr.HomeDir, DefaultKubeConfig)
	}
	return file
}
