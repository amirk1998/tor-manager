package tor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	bridgeDBMoatURL     = "https://bridges.torproject.org/moat/fetch"
	bridgeDBMoatVersion = "0.1.0"
	bridgeDBTimeout     = 30 * time.Second
	bridgeDBMaxBodySize = 64 * 1024 // 64 KB — prevent response inflation attacks
)

// moatFetchRequest is the JSON payload for the BridgeDB MOAT API.
type moatFetchRequest struct {
	Data []moatFetchData `json:"data"`
}

type moatFetchData struct {
	Version   string   `json:"version"`
	Type      string   `json:"type"`
	Transport []string `json:"transport"`
}

// moatFetchResponse is the JSON response from the BridgeDB MOAT API.
type moatFetchResponse struct {
	Data []struct {
		Type    string   `json:"type"`
		Bridges []string `json:"bridges"`
		Version string   `json:"version"`
	} `json:"data"`
	Errors []struct {
		Code   int    `json:"code"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

// BridgeDBOptions configures the BridgeDB request.
type BridgeDBOptions struct {
	// SocksPort is the local Tor SOCKS5 port to route the request through.
	// Set to 0 to make a direct request (useful if Tor isn't running yet).
	SocksPort int
	// Transport is the desired pluggable transport type.
	Transport TransportType
	// Count is the number of bridges to request (1–10, default 3).
	Count int
}

// RequestBridgesFromBridgeDB fetches fresh bridge lines from the Tor Project's
// BridgeDB MOAT service. If opts.SocksPort > 0, the request is routed through
// the local Tor SOCKS5 proxy for privacy.
func RequestBridgesFromBridgeDB(ctx context.Context, opts BridgeDBOptions) ([]Bridge, error) {
	if opts.Count <= 0 {
		opts.Count = 3
	}
	if opts.Count > 10 {
		opts.Count = 10
	}

	transport := string(opts.Transport)
	if transport == "" {
		transport = string(TransportObfs4)
	}

	client, err := buildHTTPClient(opts.SocksPort)
	if err != nil {
		return nil, fmt.Errorf("tor: bridgedb client: %w", err)
	}

	payload := moatFetchRequest{
		Data: []moatFetchData{
			{
				Version:   bridgeDBMoatVersion,
				Type:      "client-transports",
				Transport: []string{transport},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("tor: bridgedb marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, bridgeDBMoatURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("tor: bridgedb request: %w", err)
	}
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.api+json")
	// Mimic Tor Browser's user agent to blend in
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:102.0) Gecko/20100101 Firefox/102.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tor: bridgedb request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tor: bridgedb returned HTTP %d", resp.StatusCode)
	}

	// Read with size cap to prevent memory exhaustion
	limited := io.LimitReader(resp.Body, bridgeDBMaxBodySize)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("tor: bridgedb read response: %w", err)
	}

	var result moatFetchResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("tor: bridgedb parse response: %w", err)
	}

	if len(result.Errors) > 0 {
		msgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			msgs[i] = fmt.Sprintf("code=%d detail=%s", e.Code, e.Detail)
		}
		return nil, fmt.Errorf("tor: bridgedb error: %s", strings.Join(msgs, "; "))
	}

	var bridges []Bridge
	for _, d := range result.Data {
		for _, line := range d.Bridges {
			b, err := ParseBridgeLine(line)
			if err == nil {
				bridges = append(bridges, b)
			}
		}
	}

	if len(bridges) == 0 {
		return nil, ErrNoBridgesReturned
	}

	// Return only up to requested count
	if len(bridges) > opts.Count {
		bridges = bridges[:opts.Count]
	}

	return bridges, nil
}

// buildHTTPClient creates an HTTP client, optionally routed through a SOCKS5 proxy.
func buildHTTPClient(socksPort int) (*http.Client, error) {
	transport := &http.Transport{
		// Enforce timeouts at the transport level
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		// Disable keep-alives for one-shot bridge requests
		DisableKeepAlives: true,
	}

	if socksPort > 0 {
		proxyAddr := fmt.Sprintf("socks5://127.0.0.1:%d", socksPort)
		proxyURL, err := url.Parse(proxyAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid SOCKS5 proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   bridgeDBTimeout,
	}, nil
}
