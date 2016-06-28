// netstorage project netstorage.go
package netstorage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	//"errors"
	"bytes"
	"fmt"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"path"
	"time"
)

// A NetstorageClient uploads to Akamai Netstorage using HTTP:
// https://control.akamai.com/dl/customers/NS/NS_http_api_FS.pdf
// (login required)
type NetstorageClient struct {
	Host              string
	Folder            string
	NetstorageKeyName string
	NetstorageSecret  string
}

type NSFile struct {
	Type  string `xml:"type,attr"`
	Name  string `xml:"name,attr"`
	Size  int    `xml:"size,attr"`
	Md5   string `xml:"md5,attr"`
	Mtime uint32 `xml:"mtime,attr"`
}

type Stat struct {
	Dirctory string `xml:"directory,attr"`
	File     NSFile `xml:"file"`
}

type DuInfo struct {
	Files string `xml:"files,attr"`
	Bytes string `xml:"bytes,attr"`
}

type Du struct {
	Directory string `xml:"directory,attr"`
	Info      DuInfo `xml:"du-info"`
}

type NSError struct {
	Status  int
	Message string
}

func (e *NSError) Error() string {
	return fmt.Sprintf("%d - %s", e.Status, e.Message)
}

func NewClient(host, folder, keyname, key string) *NetstorageClient {
	nsclient := &NetstorageClient{
		Host:              host,
		Folder:            folder,
		NetstorageKeyName: keyname,
		NetstorageSecret:  key,
	}
	return nsclient
}

func (client *NetstorageClient) auth(httpRequest *http.Request, id string, filename string, unixTime int64, actionName string) {
	action := fmt.Sprintf("version=1&action=%s", actionName)
	fmt.Println("action:", action)
	httpRequest.Header.Set("X-Akamai-ACS-Action", action)
	authData := fmt.Sprintf("5, 0.0.0.0, 0.0.0.0, %d, %s, %s", unixTime, id, client.NetstorageKeyName)
	httpRequest.Header.Set("X-Akamai-ACS-Auth-Data", authData)
	hash := hmac.New(sha256.New, []byte(client.NetstorageSecret))
	fmt.Fprintf(hash, "%s/%s\nx-akamai-acs-action:%s\n", authData, filename, action)
	httpRequest.Header.Set("X-Akamai-ACS-Auth-Sign", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
}

func (client *NetstorageClient) PutReader(key string, r io.Reader, contentType string) error {
	filename := path.Join(client.Folder, key)
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s/%s", client.Host, filename), r)
	if err != nil {
		return err
	}
	client.auth(req, filename, filename, time.Now().Unix(), "upload")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		dump, _ := httputil.DumpResponse(resp, true)
		return fmt.Errorf("unexpected response code %d when uploading %s. Here's a dump of the response:\n%s", resp.StatusCode, filename, string(dump))
	}
	fmt.Printf("output: put %s", filename)
	return nil
}

func (client *NetstorageClient) Delete(key string) error {
	filename := path.Join(client.Folder, key)
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s/%s", client.Host, filename), nil)
	if err != nil {
		return err
	}
	client.auth(req, filename, filename, time.Now().Unix(), "delete")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		dump, _ := httputil.DumpResponse(resp, true)
		return fmt.Errorf("unexpected response code %d when uploading %s. Here's a dump of the response:\n%s", resp.StatusCode, filename, string(dump))
	}
	fmt.Printf("output: delete %s", filename)
	return nil
}

func (client *NetstorageClient) Dir(filepath string) (Stat, error) {
	var stat Stat
	filename := path.Join(client.Folder, filepath)

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/%s", client.Host, filename), nil)
	if err != nil {
		return stat, err
	}
	client.auth(req, filename, filename, time.Now().Unix(), "dir&format=xml")

	dump1, _ := httputil.DumpRequest(req, true)
	fmt.Println("req:")
	fmt.Println(string(dump1))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return stat, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return stat, getErrorDetails(resp)
	}

	buff, _ := getResponse(resp)

	reader := bytes.NewReader(buff)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err1 := decoder.Decode(&stat)
	if err1 != nil {
		return stat, err1
	}

	return stat, nil
}

func (client *NetstorageClient) DiskUsage(filepath string) (Du, error) {
	var nsdu Du
	filename := path.Join(client.Folder, filepath)

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/%s", client.Host, filename), nil)
	if err != nil {
		return nsdu, err
	}
	client.auth(req, filename, filename, time.Now().Unix(), "du&format=xml")

	dump1, _ := httputil.DumpRequest(req, true)
	fmt.Println("req:")
	fmt.Println(string(dump1))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nsdu, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nsdu, getErrorDetails(resp)
	}

	buff, _ := getResponse(resp)

	reader := bytes.NewReader(buff)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err1 := decoder.Decode(&nsdu)
	if err1 != nil {
		return nsdu, err1
	}

	return nsdu, nil
}

func (client *NetstorageClient) Statistics(filepath string) (Stat, error) {
	var stat Stat
	filename := path.Join(client.Folder, filepath)

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/%s", client.Host, filename), nil)
	if err != nil {
		return stat, err
	}
	client.auth(req, filename, filename, time.Now().Unix(), "stat&format=xml")

	dump1, _ := httputil.DumpRequest(req, true)
	fmt.Println("req:")
	fmt.Println(string(dump1))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return stat, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return stat, getErrorDetails(resp)
	}

	buff, _ := getResponse(resp)

	reader := bytes.NewReader(buff)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err1 := decoder.Decode(&stat)
	if err1 != nil {
		return stat, err1
	}

	return stat, nil
}

func getResponse(response *http.Response) ([]byte, error) {
	defer response.Body.Close()
	out, err := ioutil.ReadAll(response.Body)
	if err == io.EOF {
		err = nil
	}
	return out, err
}

func errorFromResponse(body []byte, code int) (*NSError, error) {
	var responseError NSError
	responseError.Message = string(body[:])

	responseError.Status = code
	return &responseError, nil
}

func getErrorDetails(response *http.Response) error {
	body, err := getResponse(response)
	if err != nil {
		return err
	}
	nserror, _ := errorFromResponse(body, response.StatusCode)
	return nserror
}
