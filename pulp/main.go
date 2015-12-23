// main
package main

import (
	"github.com/pulpclient/common"
	"fmt"
)

func main() {
	pc := common.NewClient("<pulp-server-url","","", "<user-name>", "<password>")	
	err := pc.Authenticate()
	
	fmt.Println(err)
	fmt.Println(pc.Cert.PkiKey)