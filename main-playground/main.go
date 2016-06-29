// main
package main

import (
	"fmt"
	"github.com/ap/netstorage"
	"github.com/ap/pulp"
)

func main() {
	ns := netstorage.NewClient("host", "base-folder", "key-name", "key")
	fmt.Println("Dir")
	stat, err := ns.Dir("/test/")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(stat)
	}
	fmt.Println("===========================")
	fmt.Println("Du")
	nsdu, err := ns.DiskUsage("")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(nsdu)
	}
	fmt.Println("===========================")
	fmt.Println("Stat")
	stat1, err := ns.Statistics("")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(stat1)
	}
	fmt.Println("============================")
	fmt.Println("MkDir")
	err1 := ns.MakeDir("/test/test1")
	fmt.Println("error", err1)

	fmt.Println("============================")
	fmt.Println("RmDir")
	err2 := ns.RemoveDir("/test/test1")
	fmt.Println("error", err2)
}

func testPulpClient() {
	pc := pulp.PulpClient("plp-server-url", "", "", "user", "passwd")

	var repos pulp.Repositories
	repos, _ = pc.ListRepositories()
	err := pc.Authenticate()
	fmt.Println("error in auth: ", err)
	fmt.Println("received key:", pc.Cert.PkiKey)
	fmt.Println("==================================================================")
	for _, repo := range repos {
		fmt.Println("url :", repo.URL)
		fmt.Println("display name:", repo.Display)
		fmt.Println("repo id:", repo.RepoId)
		fmt.Println("description: ", repo.Description)
		fmt.Println("importers: ", repo.Importers)
		fmt.Println("distributors: ", repo.Distributors)
		fmt.Println("------------------------------")
	}

	fmt.Println("==================================================================")
	fmt.Println("retrieving details of repository: redhat-rhel7-docker-hello-world ")
	var repository pulp.RepositoryDetails
	repository, _ = pc.GetRepository("redhat-rhel7-docker-hello-world")
	fmt.Println("url: ", repository.URL)
	fmt.Println("display name: ", repository.Display)

	//	fmt.Println("======================================================================")
	//	fmt.Println("Creating a new repository")
	//	var createrepo pulp.RepositoryDetails
	//	createrepo.Description = ""
	//	createrepo.RepoId = "hello-go-1"

	//	repoc, err := pc.CreateRepository(createrepo)
	//	fmt.Println("error, if any", err)
	//	fmt.Printf(repoc.RepoId)

	fmt.Println("=======================================================================")
	fmt.Println("Listing all upload requests")
	var uploadReqs pulp.UploadRequests
	uploadReqs, _ = pc.ListUploadRequests()
	fmt.Println("Uplaoad requests: ", uploadReqs.UploadIds)

	//	fmt.Println("=======================================================================")
	//	fmt.Println("Creating an upload request")
	//	var uploadReq pulp.UploadRequest
	//	uploadReq, _ = pc.CreateUploadRequest()
	//	fmt.Println("Uplaoad request ID ", uploadReq.UploadId)
}
