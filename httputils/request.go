package httputils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Post return (statusCode, body, error)
func Post(ctx context.Context, url string, body interface{}, headers ...map[string]string) (int, []byte, error) {
	js, err := json.Marshal(body)
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(js))
	if err != nil {
		return 0, nil, err
	}

	return perform(ctx, req, headers...)
}

// Get return (statusCode, body, error)
func Get(ctx context.Context, url string, query map[string]interface{}, headers ...map[string]string) (int, []byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, err
	}

	if len(query) > 0 {
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
	}

	return perform(ctx, req, headers...)
}

func perform(ctx context.Context, req *http.Request, headers ...map[string]string) (status int, body []byte, err error) {
	req = req.WithContext(ctx)

	// Set converts key to canonicalMIMEkey X-Api-Token
	// so no need to transform key
	if len(headers) > 0 {
		for k, v := range headers[0] {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	// An error is returned if caused by client policy (such as CheckRedirect),
	// or failure to speak HTTP (such as a network connectivity problem).
	// A non-2xx status code doesn't cause an error.
	// On error, any Response can be ignored. A non-nil Response with a non-nil
	// error only occurs when CheckRedirect fails, and even then the returned
	// Response.Body is already closed.
	if err != nil {
		if resp != nil {
			return resp.StatusCode, nil, err
		}
		return http.StatusInternalServerError, nil, err
	}

	// We will close only when no error as response is always closed in error case
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return http.StatusNoContent, nil, err
	}

	return resp.StatusCode, b, nil
}
