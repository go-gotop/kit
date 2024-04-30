package bnhttp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bitly/go-simplejson"
	jsoniter "github.com/json-iterator/go"
)

var (
	TimestampKey  = "timestamp"
	SignatureKey  = "signature"
	RecvWindowKey = "recvWindow"
)

// Redefining the standard package
var Json = jsoniter.ConfigCompatibleWithStandardLibrary

func currentTimestamp() int64 {
	return FormatTimestamp(time.Now())
}

// FormatTimestamp formats a time into Unix timestamp in milliseconds, as requested by Binance.
func FormatTimestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func NewJSON(data []byte) (j *simplejson.Json, err error) {
	j, err = simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// NewClient initialize an API client instance with API key and secret key.
// You should always call this function before using this SDK.
// Services will be created by the form client.NewXXXService().
func NewClient(ops ...Option) *Client {
	opts := &options{
		httpClient: http.DefaultClient,
	}
	for _, o := range ops {
		o(opts)
	}
	if opts.proxyUrl != "" {
		proxy, err := url.Parse(opts.proxyUrl)
		if err != nil {
			panic(err)
		}
		tr := &http.Transport{
			Proxy: http.ProxyURL(proxy),
			// TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		opts.httpClient.Transport = tr
	}
	return &Client{
		userAgent: "GoTop",
		opts:      opts,
	}
}

// APIError define API error when response status is 4xx or 5xx
type APIError struct {
	Code    int64  `json:"code"`
	Message string `json:"msg"`
}

// Error return error code and message
func (e APIError) Error() string {
	return fmt.Sprintf("<APIError> code=%d, msg=%s", e.Code, e.Message)
}

// IsAPIError check if e is an API error
func IsAPIError(e error) bool {
	_, ok := e.(*APIError)
	return ok
}

type doFunc func(req *http.Request) (*http.Response, error)

// Client define API client
type Client struct {
	baseURL   string
	opts      *options
	userAgent string
	do        doFunc
}

func (c *Client) parseRequest(r *Request, opts ...RequestOption) (err error) {
	// set request options from user
	for _, opt := range opts {
		opt(r)
	}
	err = r.validate()
	if err != nil {
		return err
	}

	fullURL := fmt.Sprintf("%s%s", c.baseURL, r.Endpoint)
	if r.recvWindow > 0 {
		r.SetParam(RecvWindowKey, r.recvWindow)
	}
	if r.SecType == SecTypeSigned {
		r.SetParam(TimestampKey, currentTimestamp()-c.opts.timeOffset)
	}
	queryString := r.query.Encode()
	body := &bytes.Buffer{}
	bodyString := r.form.Encode()
	header := http.Header{}
	if r.header != nil {
		header = r.header.Clone()
	}
	if bodyString != "" {
		header.Set("Content-Type", "application/x-www-form-urlencoded")
		body = bytes.NewBufferString(bodyString)
	}
	if r.SecType == SecTypeAPIKey || r.SecType == SecTypeSigned {
		header.Set("X-MBX-APIKEY", r.APIKey)
	}

	if r.SecType == SecTypeSigned {
		raw := fmt.Sprintf("%s%s", queryString, bodyString)
		mac := hmac.New(sha256.New, []byte(r.SecretKey))
		_, err = mac.Write([]byte(raw))
		if err != nil {
			return err
		}
		v := url.Values{}
		v.Set(SignatureKey, fmt.Sprintf("%x", (mac.Sum(nil))))
		if queryString == "" {
			queryString = v.Encode()
		} else {
			queryString = fmt.Sprintf("%s&%s", queryString, v.Encode())
		}
	}
	if queryString != "" {
		fullURL = fmt.Sprintf("%s?%s", fullURL, queryString)
	}

	r.fullURL = fullURL
	r.header = header
	r.body = body
	return nil
}

func (c *Client) CallAPI(ctx context.Context, r *Request, opts ...RequestOption) (data []byte, err error) {
	err = c.parseRequest(r, opts...)
	if err != nil {
		return []byte{}, err
	}
	req, err := http.NewRequest(r.Method, r.fullURL, r.body)
	if err != nil {
		return []byte{}, err
	}
	req = req.WithContext(ctx)
	req.Header = r.header
	f := c.do
	if f == nil {
		f = c.opts.httpClient.Do
	}
	res, err := f(req)
	if err != nil {
		return []byte{}, err
	}
	data, err = io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		cerr := res.Body.Close()
		// Only overwrite the retured error if the original error was nil and an
		// error occurred while closing the body.
		if err == nil && cerr != nil {
			err = cerr
		}
	}()

	if res.StatusCode >= http.StatusBadRequest {
		apiErr := new(APIError)
		e := Json.Unmarshal(data, apiErr)
		if e != nil {
			return nil, e
		}
		return nil, apiErr
	}
	return data, nil
}

// SetApiEndpoint set api Endpoint
func (c *Client) SetApiEndpoint(url string) {
	c.baseURL = url
}
