package core

import (
	"io"
	"net/http"
	"net/url"
)

type GoPreparedHttpReq struct {
	url *url.URL
	method string
	body io.Reader
	req *http.Request
	client *http.Client
}
