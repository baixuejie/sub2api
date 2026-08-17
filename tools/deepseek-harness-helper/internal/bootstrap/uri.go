package bootstrap

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var ErrInvalidLaunchURI = errors.New("invalid bootstrap URI")

func ParseLaunchURI(raw string) (LaunchRequest, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !strings.EqualFold(u.Scheme, "sub2api-harness") || !strings.EqualFold(u.Host, "bootstrap") || u.User != nil || u.Fragment != "" {
		return LaunchRequest{}, ErrInvalidLaunchURI
	}
	if u.Path != "" && u.Path != "/" {
		return LaunchRequest{}, ErrInvalidLaunchURI
	}
	q := u.Query()
	if len(q) != 3 || len(q["server"]) != 1 || len(q["ticket"]) != 1 || len(q["operation_id"]) != 1 {
		return LaunchRequest{}, ErrInvalidLaunchURI
	}
	server, err := ValidateServerURL(q.Get("server"))
	if err != nil {
		return LaunchRequest{}, err
	}
	ticket := strings.TrimSpace(q.Get("ticket"))
	operationID := strings.TrimSpace(q.Get("operation_id"))
	if !validOpaque(ticket, 4096) || !validOpaque(operationID, 128) {
		return LaunchRequest{}, ErrInvalidLaunchURI
	}
	return LaunchRequest{Server: server.String(), Ticket: ticket, OperationID: operationID}, nil
}

func ValidateServerURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("%w: invalid server", ErrInvalidLaunchURI)
	}
	if u.Path != "" && u.Path != "/" {
		return nil, fmt.Errorf("%w: server must be an origin", ErrInvalidLaunchURI)
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if u.Scheme == "https" {
		return &url.URL{Scheme: "https", Host: u.Host}, nil
	}
	if u.Scheme == "http" && isLoopbackHost(host) {
		return &url.URL{Scheme: "http", Host: u.Host}, nil
	}
	return nil, fmt.Errorf("%w: server must use HTTPS or localhost HTTP", ErrInvalidLaunchURI)
}

func ValidateStatusURL(raw, serverOrigin, operationID string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("invalid status_url")
	}
	server, err := ValidateServerURL(serverOrigin)
	if err != nil || !strings.EqualFold(u.Scheme, server.Scheme) || !strings.EqualFold(u.Host, server.Host) {
		return nil, errors.New("status_url must use the exchange server origin")
	}
	expected := "/api/v1/deepseek-harness/sessions/" + url.PathEscape(operationID) + "/events"
	if u.EscapedPath() != expected {
		return nil, errors.New("status_url path does not match operation_id")
	}
	return u, nil
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func validOpaque(value string, max int) bool {
	if value == "" || len(value) > max {
		return false
	}
	for _, r := range value {
		if r <= 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}
