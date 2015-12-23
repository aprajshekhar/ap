// netstorage project netstorage.go
package netstorage

import (
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const version = "1"

type httpError struct {
	error
	code int
}

func NewHTTPError(resp *http.Response) *httpError {
	code := resp.StatusCode
	return &httpError{
		error: errors.New(http.StatusText(code)),
		code:  code,
	}
}

func NewHTTPErrorWithText(resp *http.Response, txt string) *httpError {
	code := resp.StatusCode
	return &httpError{
		error: errors.New(http.StatusText(code) + " - " + txt),
		code:  code,
	}
}

// Api instances are safe for concurrent use by multiple goroutines
type NetStorage struct {
	client  	*http.Client
	KeyName 	string
	Secret  	string
	HostName	string
	UseSSL		bool
}

func NewNetStorageDefault(keyName, secret, hostName string) NetStorage {
	
	return NetStorage{keyName, secret, hostName, false}
}

func NewNetStorage(keyName, secret, hostName string, useSSL bool){
	client := &http.Client{}
	return Api{client, keyName, secret, hostName, useSSL}
}

func (netStorage NetStorage)SetKeyName(keyName string){
	netStorage.KeyName = keyName
}

func (netStorage NetStorage)SetSecret(secret string){
	netStorage.Secret = secret
}

func (netStorage NetStorage)SetHostName(hostName string){
	netStorage.HostName = hostName
}

func (netStorage NetStorage)SetUseSSL(useSSL bool){
	netStorage.UseSSL = useSSL
}

func (netStorage NetStorage)NetStorageUrl(path string) string{
	scheme := "http"
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if netStorage.UseSSL {
		scheme = "https"
	}

	host := fmt.Sprintf("%s.%s", netStorage.HostName, path)

	netStorageUrl := &url.URL{
		Scheme: scheme,
		Host:   host}
	return netStorageUrl.String()
}

func (netStorage NetStorage) auth(req *http.Request, rel_path, action string) {
	data, signature := api.sign(rel_path, action, -1, -1)
	req.Header.Add("X-Akamai-ACS-Auth-Data", data)
	req.Header.Add("X-Akamai-ACS-Auth-Sign", signature)
}

