package client

import (
	"github.com/dmitrygulevich2000/tiny-redis-cache/api"

	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Client interface {
	URL(ep string) *url.URL
	Do(r *http.Request) (*http.Response, []byte, error)
}

// non-positive timeout means no timeout
func NewClient(address string, timeout time.Duration) (Client, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimRight(u.Path, "/")

	if timeout < 0 {
		timeout = 0
	}
	c := &httpClient{
		endpoint: u,
		client: http.Client {
			Timeout: timeout,
		},
	}

	return c, nil
}

type httpClient struct {
	endpoint *url.URL
	client http.Client
}

func (c *httpClient) URL(ep string) *url.URL {
	p := path.Join(c.endpoint.Path, ep)

	u := *c.endpoint
	u.Path = p

	return &u
}

func (c *httpClient) Do(r *http.Request) (*http.Response, []byte, error) {
	resp, err := c.client.Do(r)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, body, err
}


type ClientAPI interface {
	// return value: "OK"
	Set(key string, value interface{}, ttl time.Duration) (interface{}, error)
	Get(key string) (interface{}, error)
	Del(keys ...string) (int, error)
	// Keys(pattern string) ([]string, error)
}

func NewAPI(c Client) ClientAPI {
	return &httpAPI{
		client: c,
	}
}

type httpAPI struct {
	client Client
}

func (h *httpAPI) Set(key string, value interface{}, ttl time.Duration) (interface{}, error) {
	params := &api.SetParams {
		Key: key,
		Value: value,
		Ttl: ttl,
	}
	reqBody, err := json.Marshal(params) 
	if err != nil {
		return nil, err
	}

	url := h.client.URL("/set")
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header["Content-Type"] = []string{"application/json"}

	_, body, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	var result interface{}
	return result, json.Unmarshal(body, &result)
}

func (h *httpAPI) Get(key string) (interface{}, error) {
	params := &api.GetParams {
		Key: key,
	}
	reqBody, err := json.Marshal(params) 
	if err != nil {
		return nil, err
	}

	url := h.client.URL("/get")
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header["Content-Type"] = []string{"application/json"}

	_, body, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	var result interface{}
	return result, json.Unmarshal(body, &result)
}

func (h *httpAPI) Del(keys ...string) (int, error) {
	params := &api.DelParams {
		Keys: keys,
	}
	reqBody, err := json.Marshal(params) 
	if err != nil {
		return 0, err
	}

	url := h.client.URL("/del")
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return 0, err
	}
	req.Header["Content-Type"] = []string{"application/json"}

	_, body, err := h.client.Do(req)
	if err != nil {
		return 0, err
	}

	var result int
	return result, json.Unmarshal(body, &result)
}