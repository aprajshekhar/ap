//
package pulp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	mux    *http.ServeMux
	client *Client
	server *httptest.Server
)

func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	fmt.Println("server", server.URL)
	// pulp client configured to use test server
	client = PulpClient("", "", "", "test", "test")
	client.Endpoint = server.URL
}

func TestAuthenticate(t *testing.T) {
	var err error

	expectedCert := Certificate{PkiCertificate: "pkiCert", PkiKey: "key"}
	expectedJson, _ := json.Marshal(expectedCert)
	setup()
	defer teardown()

	mux.HandleFunc("/pulp/api/v2/actions/login/",
		func(w http.ResponseWriter, r *http.Request) {
			checkMethod(t, r, "POST")
			fmt.Fprint(w, string(expectedJson[:]))
		},
	)
	if err = client.Authenticate(); err != nil {
		t.Errorf("API error: %s", err)
	}
	if !reflect.DeepEqual(client.Cert, expectedCert) {
		t.Errorf("got %#v expected %#v", client.Cert, expectedCert)
	}
}

func TestListRepositories(t *testing.T) {
	var repolist, expected Repositories
	var err error
	expectedRepoDetails := RepositoryDetails{URL: "/pulp/api/v2/repositories/test", RepoId: "test", Description: "test repo", Display: "test-unit"}
	expected = make(Repositories, 1)
	expected[0] = expectedRepoDetails
	expectedJson, _ := json.Marshal(expected)
	setup()
	defer teardown()

	mux.HandleFunc("/pulp/api/v2/repositories/",
		func(w http.ResponseWriter, r *http.Request) {
			checkMethod(t, r, "GET")
			fmt.Fprint(w, string(expectedJson[:]))
		},
	)
	if repolist, err = client.ListRepositories(); err != nil {
		t.Errorf("API error: %s", err)
	}
	if !reflect.DeepEqual(repolist, expected) {
		t.Errorf("got %#v expected %#v", repolist, expected)
	}
}

func TestGetRepository(t *testing.T) {
	var repo RepositoryDetails
	var err error
	expectedRepoDetails := RepositoryDetails{URL: "/pulp/api/v2/repositories/test", RepoId: "test", Description: "test repo", Display: "test-unit"}

	expectedJson, _ := json.Marshal(expectedRepoDetails)
	setup()
	defer teardown()

	mux.HandleFunc("/pulp/api/v2/repositories/test/",
		func(w http.ResponseWriter, r *http.Request) {
			checkMethod(t, r, "GET")
			fmt.Fprint(w, string(expectedJson[:]))
		},
	)
	if repo, err = client.GetRepository("test"); err != nil {
		t.Errorf("API error: %s", err)
	}
	if !reflect.DeepEqual(repo, expectedRepoDetails) {
		t.Errorf("got %#v expected %#v", repo, expectedRepoDetails)
	}
}

func TestCreateRepo(t *testing.T) {
	var err error

	var createrepo, recievedRepo RepositoryDetails
	createrepo.Description = "test repo"
	createrepo.RepoId = "test"
	createrepo.Display = "test-unit"
	createrepo.URL = "/pulp/api/v2/repositories/test"

	expectedRepoDetails := RepositoryDetails{URL: "/pulp/api/v2/repositories/test", RepoId: "test", Description: "test repo", Display: "test-unit"}
	expectedJson, _ := json.Marshal(expectedRepoDetails)

	setup()
	defer teardown()

	mux.HandleFunc("/pulp/api/v2/repositories/",
		func(w http.ResponseWriter, r *http.Request) {
			checkMethod(t, r, "POST")
			fmt.Fprint(w, string(expectedJson[:]))
		},
	)
	if recievedRepo, err = client.CreateRepository(createrepo); err != nil {
		t.Errorf("API error: %s", err)
	}
	if !reflect.DeepEqual(recievedRepo, expectedRepoDetails) {
		t.Errorf("got %#v expected %#v", recievedRepo, expectedRepoDetails)
	}
}

func TestListUploadRequests(t *testing.T) {
	var uploadreqlist, expected UploadRequests
	var err error
	expected = UploadRequests{[]string{"abc123", "def456"}}
	expectedJson, _ := json.Marshal(expected)
	setup()
	defer teardown()

	mux.HandleFunc("/pulp/api/v2/content/uploads/",
		func(w http.ResponseWriter, r *http.Request) {
			checkMethod(t, r, "GET")
			fmt.Fprint(w, string(expectedJson[:]))
		},
	)
	if uploadreqlist, err = client.ListUploadRequests(); err != nil {
		t.Errorf("API error: %s", err)
	}
	if !reflect.DeepEqual(uploadreqlist, expected) {
		t.Errorf("got %#v expected %#v", uploadreqlist, expected)
	}
}

func TestCreateUploadRequest(t *testing.T) {
	var uploadreq, expected UploadRequest
	var err error
	expected = UploadRequest{Href: "/pulp/api/v2/repositories/test", UploadId: "abc123"}
	expectedJson, _ := json.Marshal(expected)
	setup()
	defer teardown()

	mux.HandleFunc("/pulp/api/v2/content/uploads/",
		func(w http.ResponseWriter, r *http.Request) {
			checkMethod(t, r, "POST")
			fmt.Fprint(w, string(expectedJson[:]))
		},
	)
	if uploadreq, err = client.CreateUploadRequest(); err != nil {
		t.Errorf("API error: %s", err)
	}
	if !reflect.DeepEqual(uploadreq, expected) {
		t.Errorf("got %#v expected %#v", uploadreq, expected)
	}
}

func checkMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func teardown() {
	server.Close()
}
