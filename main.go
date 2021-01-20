/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 08.12.2020
 * Last Modified Date: 19.01.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package main

import (
	"os"

	podManager "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/manager"
	server "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/server"
)

func main() {

	goEnv := os.Getenv("GOENV")
	var inClusterConfig bool
	if goEnv == "" {
		panic("GOENV environement is not set !")
	} else if goEnv != "production" {
		inClusterConfig = false
	} else {
		inClusterConfig = true
	}
	manager := podManager.New(inClusterConfig)
	server.Listen(*manager)
}
