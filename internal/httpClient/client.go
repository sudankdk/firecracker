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
		DialContext: func(ctx context.Context, _ string, _ string) (net.Conn, error) {
			d := &net.Dialer{}
			return d.DialContext(ctx, "unix", sock)
		},
	}
	return &http.Client{Transport: transport, Timeout: 10 * time.Second}
}

func Put(client *http.Client, path string, body []byte) error {
	req, _ := http.NewRequest(
		"PUT",
		"http://localhost"+path,
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("firecracker error: %s", resp.Status)
	}
	return nil
}
