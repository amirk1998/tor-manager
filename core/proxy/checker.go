// Package proxy provides utilities for verifying connectivity through a Tor SOCKS5 proxy.
package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	ipCheckURL      = "https://check.torproject.org/api/ip"
	ipInfoURL       = "https://ipinfo.io/json"
	defaultTimeout  = 20 * time.Second
	maxResponseSize = 8 * 1024 // 8 KB
)

// IPInfo holds information about the current exit IP address.
type IPInfo struct {
	IP       string `json:"ip"`
	IsTor    bool   `json:"IsTor"` // from check.torproject.org
	Country  string `json:"country"`
	Region   string `json:"region"`
	City     string `json:"city"`
	Org      string `json:"org"` // ASN + organization name
	Hostname string `json:"hostname"`
}

// CheckResult combines Tor connectivity status with IP details.
type CheckResult struct {
	*IPInfo
	// Latency is the round-trip time for the IP check request.
	Latency time.Duration
	// Error is non-nil if the check failed.
	Error error
}

// Checker performs connectivity and IP checks through a Tor SOCKS5 proxy.
type Checker struct {
	socksAddr string // e.g. "127.0.0.1:9050"
	client    *http.Client
}

// NewChecker creates a Checker routed through the given SOCKS5 port.
func NewChecker(socksPort int) (*Checker, error) {
	if socksPort < 1 || socksPort > 65535 {
		return nil, fmt.Errorf("proxy: invalid SOCKS5 port: %d", socksPort)
	}

	socksAddr := fmt.Sprintf("127.0.0.1:%d", socksPort)
	proxyURL, err := url.Parse("socks5://" + socksAddr)
	if err != nil {
		return nil, fmt.Errorf("proxy: invalid SOCKS5 URL: %w", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 0, // no keep-alive for privacy (each request = new circuit potential)
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		DisableKeepAlives:     true, // force new TCP connection per request
	}

	return &Checker{
		socksAddr: socksAddr,
		client: &http.Client{
			Transport: transport,
			Timeout:   defaultTimeout,
		},
	}, nil
}

// CheckIP fetches the current exit IP and verifies it's a Tor exit node.
// It first calls check.torproject.org (authoritative), then enriches
// with ipinfo.io for geolocation.
func (c *Checker) CheckIP(ctx context.Context) CheckResult {
	start := time.Now()

	// --- Step 1: check.torproject.org (tells us if we're actually using Tor) ---
	type torCheckResp struct {
		IP    string `json:"IP"`
		IsTor bool   `json:"IsTor"`
	}

	var torInfo torCheckResp
	if err := c.fetchJSON(ctx, ipCheckURL, &torInfo); err != nil {
		return CheckResult{
			Error:   fmt.Errorf("proxy: tor check failed: %w", err),
			Latency: time.Since(start),
		}
	}

	info := &IPInfo{
		IP:    torInfo.IP,
		IsTor: torInfo.IsTor,
	}

	// --- Step 2: ipinfo.io for geolocation (best-effort) ---
	type ipInfoResp struct {
		IP       string `json:"ip"`
		Country  string `json:"country"`
		Region   string `json:"region"`
		City     string `json:"city"`
		Org      string `json:"org"`
		Hostname string `json:"hostname"`
	}
	var geoInfo ipInfoResp
	if err := c.fetchJSON(ctx, ipInfoURL, &geoInfo); err == nil {
		info.Country = geoInfo.Country
		info.Region = geoInfo.Region
		info.City = geoInfo.City
		info.Org = geoInfo.Org
		info.Hostname = geoInfo.Hostname
	}

	return CheckResult{
		IPInfo:  info,
		Latency: time.Since(start),
	}
}

// CheckConnectivity does a lightweight check to see if the SOCKS5 proxy
// is reachable at all (doesn't make an external HTTP request).
func (c *Checker) CheckConnectivity(ctx context.Context) error {
	d := net.Dialer{Timeout: 3 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", c.socksAddr)
	if err != nil {
		return fmt.Errorf("proxy: SOCKS5 not reachable at %s: %w", c.socksAddr, err)
	}
	conn.Close()
	return nil
}

// ProxyAddress returns the SOCKS5 proxy address string (e.g. "socks5://127.0.0.1:9050").
func (c *Checker) ProxyAddress() string {
	return "socks5://" + c.socksAddr
}

// fetchJSON performs a GET request through Tor and decodes the JSON response.
func (c *Checker) fetchJSON(ctx context.Context, rawURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("proxy: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	// Avoid sending Accept-Language or other fingerprinting headers
	req.Header.Del("Accept-Encoding")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("proxy: request to %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("proxy: %s returned HTTP %d", rawURL, resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxResponseSize)
	if err := json.NewDecoder(limited).Decode(out); err != nil {
		return fmt.Errorf("proxy: decode response from %s: %w", rawURL, err)
	}
	return nil
}

// FormatIPInfo returns a compact, display-friendly summary of an IPInfo.
func FormatIPInfo(i *IPInfo) string {
	if i == nil {
		return "unknown"
	}
	loc := ""
	if i.City != "" && i.Country != "" {
		loc = fmt.Sprintf(" | %s, %s", i.City, i.Country)
	} else if i.Country != "" {
		loc = " | " + i.Country
	}
	tor := ""
	if i.IsTor {
		tor = " [Tor exit]"
	} else {
		tor = " [!NOT a Tor exit]"
	}
	return fmt.Sprintf("%s%s%s", i.IP, loc, tor)
}
