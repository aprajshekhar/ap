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

func checkMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func teardown() {
	server.Close()
}
