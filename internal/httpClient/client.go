package client

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

func NewClient(sock string) *http.Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			d := &net.Dialer{}
			return d.DialContext(ctx, "unix", sock)
		},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
}

func Put(client *http.Client, path string, body []byte) error {
	req, err := http.NewRequest(
		http.MethodPut,
		"http://localhost"+path,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("firecracker PUT %s failed: %s", path, resp.Status)
	}
	return nil
}
