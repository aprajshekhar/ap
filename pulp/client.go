// pulp project pulp.go
package pulp

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/core/http"
	"github.com/Azure/azure-sdk-for-go/core/tls"
	"io"
	"io/ioutil"
	"strconv"
)

type ContentUnitCounts struct {
	DockerBlob     int `json:"docker_blob"`
	DockerImage    int `json:"docker_image"`
	DockerManifest int `json:"docker_manifest"`
}

type Id struct {
	Oid string `json:"$oid"`
}

type Note struct {
	RepoType string `json:"_repo-type"`
}

type Filters struct {
	Unit        string `json:"unit"`
	Association string `json:"association"`
}

type ErrorDetail struct {
	description string `json:"description"`
}

type ErrorResponse struct {
	Code        string      `json:"http_status"`
	ErrorDetail ErrorDetail `json:"error"`
}

type pulpResponse struct {
	status  int
	headers http.Header
	body    []byte
}

type Client struct {
	Endpoint string
	Cert     Certificate
	UserName string //credentials to use
	Password string //if certificate auth is not being used
}

type Certificate struct {
	PkiCertificate string `json:"certificate"`
	PkiKey         string `json:"key"`
}

type Tag struct {
	ImageID string `json:"image_id,omitempty"`
	Name    string `json:"tag,omitempty"`
}
type ScratchPad struct {
	Tags []Tag `json:"tags,omitempty"`
}
type RepositoryDetails struct {
	URL             string            `json:"_href"`
	PulpId          Id                `json:"_id"`
	Ns              string            `json:"_ns"`
	Description     string            `json:"description"`
	Display         string            `json:"display_name"`
	RepoId          string            `json:"id"`
	LastUnitAdded   string            `json:"last_unit_added"`
	LastUnitRemoved string            `json:"last_unit_removed"`
	Notes           Note              `json:"notes"`
	UnitCounts      ContentUnitCounts `json:"content_unit_counts"`
	ImageDetails    ScratchPad        `json:"scratchpad,omitempty"`
}

type Repositories []RepositoryDetails

func NewClient(endpoint, pkicert, pkikey, username, password string) *Client {
	client := &Client{
		Endpoint: endpoint,
		UserName: username,
		Password: password,
		Cert: Certificate{
			PkiCertificate: pkicert,
			PkiKey:         pkikey,
		},
	}

	return client
}

// Format of reply when http error code is not 200.
// Format may be:
// {"error": "reason"}
// {"error": {"param": "reason"}}
type requestError struct {
	Error interface{} `json:"error"` // Description of this error.
}

func (client *Client) ListRepositories() (Repositories, error) {
	var pulpresponse *pulpResponse
	var err error
	var repository Repositories
	if pulpresponse, err = execute("GET", "/pulp/api/v2/repositories/", nil, client.Endpoint, client.UserName, client.Password); err != nil {
		return nil, err
	}

	marshalError := json.Unmarshal(pulpresponse.body, &repository)
	if marshalError != nil {
		return nil, marshalError
	}

	return repository, nil
}

func (client *Client) Authenticate() error {

	var cert Certificate
	var pulpresponse *pulpResponse
	fmt.Println("endpoint: " + client.Endpoint)
	var err error
	if pulpresponse, err = execute("POST", "/pulp/api/v2/actions/login/", nil, client.Endpoint, client.UserName, client.Password); err != nil {
		return err
	}

	if err = json.Unmarshal(pulpresponse.body, &cert); err != nil {
		return err
	}
	client.Cert = cert

	return nil
}

func (client *Client) GetRepository(repositoryName string) (RepositoryDetails, error) {
	var response *pulpResponse
	var err error
	var repository RepositoryDetails

	if response, err = execute("GET", "/pulp/api/v2/repositories/"+repositoryName+"/", nil, client.Endpoint, client.UserName, client.Password); err != nil {
		return repository, err
	}

	if err = json.Unmarshal(response.body, &repository); err != nil {
		return repository, err
	}
	return repository, nil
}

func errorFromJson(body []byte, code int) (*ErrorResponse, error) {
	var responseError *ErrorResponse
	if err := json.Unmarshal(body, responseError); err != nil {
		return responseError, err
	}
	responseError.Code = strconv.Itoa(code)
	return responseError, nil
}

func execute(verb, url string, content io.Reader, endPoint, userName, password string) (*pulpResponse, error) {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defaultClient := &http.Client{Transport: transport}
	request, err := http.NewRequest(verb, endPoint+url, content)

	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(userName, password)
	response, err := defaultClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	statusCode := response.StatusCode
	if statusCode >= 400 && statusCode <= 505 {
		var responseBody []byte
		responseBody, err = getResponse(response)
		if err != nil {
			return nil, err
		}
		fmt.Print(responseBody)
		if len(responseBody) == 0 {
			// no error in response body
			err = fmt.Errorf("pulp: service returned without a response body (%s)", response.Status)
		} else {
			// response contains pulp service error object, unmarshal
			errorResponse, errIn := errorFromJson(responseBody, response.StatusCode)
			if err != nil { // error unmarshaling the error response
				err = errIn
			}
			err = errorResponse
		}
		return &pulpResponse{
			status:  response.StatusCode,
			headers: response.Header,
			body:    responseBody,
		}, err
	}

	var responseBody []byte
	responseBody, err = getResponse(response)
	if err != nil {
		return nil, err
	}
	fmt.Println("response status:", response.Status)
	return &pulpResponse{
		status:  response.StatusCode,
		headers: response.Header,
		body:    responseBody,
	}, nil

	return nil, nil
}
func getResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err == io.EOF {
		err = nil
	}
	return out, err
}
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("pulp: service returned error: Code=%s, ErrorMessage=%s", e.Code, e.ErrorDetail.description)
}
