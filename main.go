package main

import (
	_ "embed"
	"gcnf/internal/cmd"
)

//go:embed project.properties
var projectPropertiesContent []byte

//go:embed client_secret.json
var clientSecretsContent []byte

func main() {
	cmd.Execute(string(projectPropertiesContent), clientSecretsContent)
}
