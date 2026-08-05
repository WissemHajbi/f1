package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Response struct {
	Body         []byte
	ContentType  string
	ETag         string
	LastModified string
	StatusCode   int
}

type Client struct {
	httpClient *http.Client
	userAgent  string
	maxBody    int64
}

func New(timeout time.Duration, userAgent string, maxBody int64) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		userAgent:  userAgent,
		maxBody:    maxBody,
	}
}

func (c *Client) Get(ctx context.Context, endpoint, accept string) (Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Response{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", accept)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("request %s: %w", endpoint, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, c.maxBody+1))
	if err != nil {
		return Response{}, fmt.Errorf("read response: %w", err)
	}
	if int64(len(body)) > c.maxBody {
		return Response{}, fmt.Errorf("response exceeds %d bytes", c.maxBody)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Response{}, fmt.Errorf("upstream returned %s", res.Status)
	}
	return Response{
		Body: body, ContentType: res.Header.Get("Content-Type"), StatusCode: res.StatusCode,
		ETag: res.Header.Get("ETag"), LastModified: res.Header.Get("Last-Modified"),
	}, nil
}
