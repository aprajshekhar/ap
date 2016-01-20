// main
package main

import (
	"fmt"
	"github.com/ap/pulp"
)

func main() {
	pc := pulp.NewClient("https://brew-pulp-docker01.web.qa.ext.phx1.redhat.com", "", "", "admin", "admin")
	//var rep pulp.Repository
	rep, _ := pc.GetRepositories()
	err := pc.Authenticate()
	fmt.Println("error in auth: ", err)
	fmt.Println(pc.Cert.PkiKey)
	fmt.Println("==================================================================")
	fmt.Println(rep)
}
