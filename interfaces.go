package main

type Response interface {
	GetStatusCode() int
	GetBody() interface{}
}

type HTTPClient interface {
	Post(url string, body interface{}) Response
	Get(url string) Response
}

type HTTPResponse struct {
	StatusCode int
	Body       interface{}
}

func (r HTTPResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r HTTPResponse) GetBody() interface{} {
	return r.Body
}

type MockHTTPClient struct{}

func (c *MockHTTPClient) Post(url string, body interface{}) Response {
	return HTTPResponse{StatusCode: 200, Body: body}
}

func (c *MockHTTPClient) Get(url string) Response {
	return HTTPResponse{StatusCode: 200, Body: "GET response"}
}