/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 08.12.2020
 * Last Modified Date: 09.12.2020
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package main

import (
	server "github.com/saisona/go-feather-slack-app/src/go-feather-slack-app/server"
)

func main() {
	server.Listen()
}
