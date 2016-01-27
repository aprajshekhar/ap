// main
package main

import (
	"fmt"
	"github.com/ap/pulp"
)

func main() {
	pc := pulp.NewClient("plp-server-url", "", "", "user", "passwd")

	var rep pulp.Repositories
	rep, _ = pc.ListRepositories()
	err := pc.Authenticate()
	fmt.Println("error in auth: ", err)
	fmt.Println("received key:", pc.Cert.PkiKey)
	fmt.Println("==================================================================")
	for _, repo := range rep {
		fmt.Println("url :", repo.URL)
		fmt.Println("display name:", repo.Display)
		fmt.Println("repo id:", repo.RepoId)
		fmt.Println("description: ", repo.Description)
		fmt.Println("------------------------------")
	}

	fmt.Println("==================================================================")
	fmt.Println("retrieving details of repository: redhat-rhel7-docker-hello-world ")
	var repository pulp.RepositoryDetails
	repository, _ = pc.GetRepository("redhat-rhel7-docker-hello-world")
	fmt.Println("url: ", repository.URL)
	fmt.Println("display name: ", repository.Display)

	fmt.Println("======================================================================")
	fmt.Println("Creating a new repository")
	var createrepo pulp.RepositoryDetails
	createrepo.Description = ""
	createrepo.Id = "hello-go-1"

	repoc, err := pc.CreateRepository(createrepo)
	fmt.Println(err)
	fmt.Printf(repoc.Id)

}
