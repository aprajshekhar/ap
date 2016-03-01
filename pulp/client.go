// pulp project pulp.go
package pulp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/core/http"
	"github.com/Azure/azure-sdk-for-go/core/tls"
	//"log"
	"io"
	"io/ioutil"
	//"strconv"
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
	Code         int      `json:"http_status"`
	ErrorMessage string   `json:"error_message"`
	Traceback    []string `json:"traceback"`
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
	URL             string            `json:"_href,omitempty"`
	PulpId          Id                `json:"_id,omitempty"`
	Ns              string            `json:"_ns,omitempty"`
	Description     string            `json:"description,omitempty"`
	Display         string            `json:"display_name,omitempty"`
	RepoId          string            `json:"id,omitempty"`
	LastUnitAdded   string            `json:"last_unit_added,omitempty"`
	LastUnitRemoved string            `json:"last_unit_removed,omitempty"`
	Notes           Note              `json:"notes,omitempty"`
	UnitCounts      ContentUnitCounts `json:"content_unit_counts,omitempty"`
	ImageDetails    ScratchPad        `json:"scratchpad,omitempty"`
	Distributors    []Distributor     `json:"distributors,omitempty"`
	Importers       []Importer        `json:"importers,omitempty"`
}

type ConfigDetail struct {
	PublishDir string `json:"publish_dir"`
	WriteFiles string `json:"write_files"`
}

type Distributor struct {
	ScratchPad        int          `json:"scratchpad"`
	Ns                string       `json:"_ns"`
	ImporterTypeId    string       `json:"importer_type_id"`
	LastPublish       string       `json:"last_publish"`
	AutoPublish       bool         `json:"auto_publish"`
	DistributorTypeId string       `json:"distributor_type_id`
	RepoId            string       `json:"repo_id"`
	PublishInProgress bool         `json:"publish_in_progress"`
	InternalId        string       `json:"_id"`
	Config            ConfigDetail `json: config`
	DistributorId     string       `json:"_id"`
}

type Importer struct {
	ScratchPad        int          `json:"scratchpad"`
	Ns                string       `json:"_ns"`
	LastPublish       string       `json:"last_publish"`
	AutoPublish       bool         `json:"auto_publish"`
	ImporterTypeId    string       `json:"importer_type_id`
	RepoId            string       `json:"repo_id,omitempty"`
	PublishInProgress bool         `json:"publish_in_progress"`
	InternalId        string       `json:"_id"`
	Config            ConfigDetail `json: config`
	ImporterId        string       `json:"_id"`
}

type UploadRequests struct {
	UploadIds []string `json:"upload_ids"`
}

type UploadRequest struct {
	Href     string `json:"_href"`
	UploadId string `json:"upload_id"`
}

type Repositories []RepositoryDetails

func PulpClient(endpoint, pkicert, pkikey, username, password string) *Client {
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

func (client *Client) CreateRepository(repodetails RepositoryDetails) (RepositoryDetails, error) {
	var response *pulpResponse
	var err error
	var repositoryResponse RepositoryDetails
	jsondata, marshalerr := json.Marshal(repodetails)

	if marshalerr != nil {
		return repositoryResponse, marshalerr
	}

	if response, err = execute("POST", "/pulp/api/v2/repositories/", jsondata, client.Endpoint, client.UserName, client.Password); err != nil {
		return repositoryResponse, err
	}

	if err = json.Unmarshal(response.body, &repositoryResponse); err != nil {
		return repositoryResponse, err
	}

	return repositoryResponse, nil

}

func (client *Client) ListUploadRequests() (UploadRequests, error) {
	var pulpresponse *pulpResponse
	var err error
	var uploadRequests UploadRequests
	if pulpresponse, err = execute("GET", "/pulp/api/v2/content/uploads/", nil, client.Endpoint, client.UserName, client.Password); err != nil {
		return uploadRequests, err
	}

	marshalError := json.Unmarshal(pulpresponse.body, &uploadRequests)
	if marshalError != nil {
		return uploadRequests, marshalError
	}

	return uploadRequests, nil
}

func (client *Client) CreateUploadRequest() (UploadRequest, error) {
	var pulpresponse *pulpResponse
	var err error
	var uploadRequest UploadRequest
	if pulpresponse, err = execute("POST", "/pulp/api/v2/content/uploads/", nil, client.Endpoint, client.UserName, client.Password); err != nil {
		return uploadRequest, err
	}

	marshalError := json.Unmarshal(pulpresponse.body, &uploadRequest)
	if marshalError != nil {
		return uploadRequest, marshalError
	}

	return uploadRequest, nil
}

func errorFromJson(body []byte, code int) (*ErrorResponse, error) {
	var responseError ErrorResponse
	if err := json.Unmarshal(body, &responseError); err != nil {
		return &responseError, err
	}

	responseError.Code = code
	return &responseError, nil
}

func execute(verb, url string, content []byte, endPoint, userName, password string) (*pulpResponse, error) {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defaultClient := &http.Client{Transport: transport}
	request, err := http.NewRequest(verb, endPoint+url, bytes.NewBuffer(content))

	if err != nil {
		return nil, err
	}

	request.SetBasicAuth(userName, password)

	response, err := defaultClient.Do(request)

	fmt.Println("Creating repo with: ", string(content[:]))
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

		if len(responseBody) == 0 {
			// no error in response body
			err = fmt.Errorf("pulp: service returned without a response body (%s)", response.Status)
		} else {
			// response contains pulp service error object, unmarshal
			errorResponse, errIn := errorFromJson(responseBody, response.StatusCode)

			if errIn != nil { // error unmarshaling the error response
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

	return &pulpResponse{
		status:  response.StatusCode,
		headers: response.Header,
		body:    responseBody,
	}, nil

	return nil, nil
}
func getResponse(response *http.Response) ([]byte, error) {
	defer response.Body.Close()
	out, err := ioutil.ReadAll(response.Body)
	if err == io.EOF {
		err = nil
	}
	return out, err
}
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("pulp: service returned error: Code=%d, ErrorMessage=%s", e.Code, e.ErrorMessage)
}
