// main
package main

import (
	"fmt"
	"github.com/ap/pulp"
)

func main() {
	pc := pulp.NewClient("url", "", "", "user", "pwd")
	//var rep pulp.Repository
	rep, _ := pc.GetRepositories()
	err := pc.Authenticate()
	fmt.Println("error in auth: ", err)
	fmt.Println(pc.Cert.PkiKey)
	fmt.Println("==================================================================")
	fmt.Println(rep)
}
