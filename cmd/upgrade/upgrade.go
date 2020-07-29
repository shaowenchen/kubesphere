package main

import (
	"kubesphere.io/kubesphere/cmd/upgrade/devops"
	"os"
)
func main() {
	cmd := devops.NewDevOpsCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
