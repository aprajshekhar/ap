// netstorage project netstorage.go
package netstorage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// A NetstorageClient uploads to Akamai Netstorage using HTTP:
// https://control.akamai.com/dl/customers/NS/NS_http_api_FS.pdf
// (login required)
type NetstorageClient struct {
	Host              string
	Folder            string
	Prefix            string
	BaseURL           string
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

func (client *NetstorageClient) SetPrefix(key string) {
	client.Prefix = key
}

func (client *NetstorageClient) URLFor(p string) string {
	return fmt.Sprintf("%s/%s.json", client.BaseURL, path.Join(client.Prefix, p))
}

func (client *NetstorageClient) auth(httpRequest *http.Request, id string, filename string, unixTime int64, actionName string) {
	action := fmt.Sprint("version=1&action=%s", actionName)
	httpRequest.Header.Set("X-Akamai-ACS-Action", action)
	authData := fmt.Sprintf("5, 0.0.0.0, 0.0.0.0, %d, %s, %s", unixTime, id, client.NetstorageKeyName)
	httpRequest.Header.Set("X-Akamai-ACS-Auth-Data", authData)
	hash := hmac.New(sha256.New, []byte(client.NetstorageSecret))
	fmt.Fprintf(hash, "%s/%s\nx-akamai-acs-action:%s\n", authData, filename, action)
	httpRequest.Header.Set("X-Akamai-ACS-Auth-Sign", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
}

func (client *NetstorageClient) PutReader(key string, r io.Reader, contentType string) error {
	filename := path.Join(client.Folder, client.Prefix, key)
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
	log.Printf("output: put %s", filename)
	return nil
}

func (client *NetstorageClient) Delete(key string) error {
	filename := path.Join(client.Folder, client.Prefix, key)
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
	log.Printf("output: delete %s", filename)
	return nil
}

func (client *NetstorageClient) Dir(key string) error {
	filename := path.Join(client.Folder, client.Prefix, key)
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/%s", client.Host, filename), nil)
	if err != nil {
		return err
	}
	client.auth(req, filename, filename, time.Now().Unix(), "dir")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		dump, _ := httputil.DumpResponse(resp, true)
		return fmt.Errorf("unexpected response code %d when uploading %s. Here's a dump of the response:\n%s", resp.StatusCode, filename, string(dump))
	}
	log.Printf("output: dir %s", filename)
	return nil
}
