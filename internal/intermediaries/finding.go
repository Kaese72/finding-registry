package intermediaries

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/Kaese72/finding-registry/rest/apierrors"
)

type ReportLocatorType string

const (
	IPv4     ReportLocatorType = "IPv4"
	Hostname ReportLocatorType = "Hostname"
	HTTP     ReportLocatorType = "HTTP"
	TCP      ReportLocatorType = "TCP"
	UDP      ReportLocatorType = "UDP"
)

const (
	// "global" is the default distinguisher for everything
	// and indicates the thing is attached to has no locality
	GlobalDistinguisher = "global"
)

func isValidIPv4(ip string) (net.IP, bool) {
	parsedIP := net.ParseIP(ip)
	return parsedIP, parsedIP != nil && parsedIP.To4() != nil
}

// Validate checks if the ReportLocator is valid together with its value.
// Returns an API Error if the validation fails.
// 400: The provided data is syntactically incorrect
// 422: The provided data is syntactically correct, but semantically incorrect, like disallowed values
func (locator ReportLocator) Validate() error {
	if locator.Type == "" {
		return apierrors.APIError{Code: http.StatusBadRequest, WrappedError: fmt.Errorf("missing Type")}
	}
	if locator.Value == "" {
		return apierrors.APIError{Code: http.StatusBadRequest, WrappedError: fmt.Errorf("missing Value")}
	}
	if locator.Distinguisher == "" {
		return apierrors.APIError{Code: http.StatusBadRequest, WrappedError: fmt.Errorf("missing Distinguisher")}
	}
	switch locator.Type {
	case IPv4:
		// Validate IPv4 address
		ip, is4 := isValidIPv4(locator.Value)
		if !is4 {
			return apierrors.APIError{Code: http.StatusBadRequest, WrappedError: fmt.Errorf("invalid IPv4 address: %s", locator.Value)}
		}
		if ip.IsPrivate() {
			// If the IP address is private, it must have a distinguisher.
			// This is to allow multiple private IPv4 addresses to be distinguished,
			// and you need to explicitly set separate or merge private addresses.
			if locator.Distinguisher == "" {
				return apierrors.APIError{Code: http.StatusUnprocessableEntity, WrappedError: fmt.Errorf("private IPv4 address must have a distinguisher")}
			}
			if locator.Distinguisher == GlobalDistinguisher {
				return apierrors.APIError{Code: http.StatusUnprocessableEntity, WrappedError: fmt.Errorf("private IPv4 address cannot have a global distinguisher")}
			}
		}
		if ip.IsLoopback() {
			// We do not allow findings to be reported on loopback addresses
			return apierrors.APIError{Code: http.StatusUnprocessableEntity, WrappedError: fmt.Errorf("loopback IPv4 address not allowed")}
		}
	// case IPv6:
	// 	// Validate IPv6 address
	// 	// FIXME global check
	case HTTP:
		// Validate URL
		_, err := url.Parse(locator.Value)
		if err != nil {
			return apierrors.APIError{Code: http.StatusBadRequest, WrappedError: fmt.Errorf("invalid URL: %s", locator.Value)}
		}
	case TCP:
		// Validate TCP address
		_, err := net.ResolveTCPAddr("tcp", locator.Value)
		if err != nil {
			return apierrors.APIError{Code: http.StatusBadRequest, WrappedError: fmt.Errorf("invalid TCP address: %s", locator.Value)}
		}
	case UDP:
		// Validate UDP address
		_, err := net.ResolveUDPAddr("udp", locator.Value)
		if err != nil {
			return apierrors.APIError{Code: http.StatusBadRequest, WrappedError: fmt.Errorf("invalid IPv4 address: %s", locator.Value)}
		}
	case Hostname:
		// Everything is a valid hostname it seems
		// FIXME there are some restrictions on hostnames after all...
		if locator.Value == "localhost" {
			return apierrors.APIError{Code: http.StatusUnprocessableEntity, WrappedError: fmt.Errorf("hostname may not be '%s'", locator.Value)}
		}
	default:
		return apierrors.APIError{Code: http.StatusBadRequest, WrappedError: fmt.Errorf("invalid ReportLocatorType: %s", locator.Type)}
	}
	return nil
}

type ReportLocator struct {
	Type          ReportLocatorType
	Value         string
	Distinguisher string
}

func (r ReportLocator) Implied() ([]ReportLocator, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	ret := []ReportLocator{r}
	switch r.Type {
	case HTTP:
		u, err := url.Parse(r.Value)
		if err != nil {
			// This should never happen, since we already validated the URL
			panic(err)
		}
		locator := ReportLocator{Type: TCP, Value: u.Host, Distinguisher: r.Distinguisher}
		if u.Port() == "" {
			if u.Scheme == "http" {
				locator.Value += ":80"
			} else if u.Scheme == "https" {
				locator.Value += ":443"
			}
		}
		downstreamLocators, err := locator.Implied()
		return append(ret, downstreamLocators...), err
	case TCP, UDP:
		// Check if the TCP/UDP address implies an IP Address
		host, _, err := net.SplitHostPort(r.Value)
		if err != nil {
			return nil, err
		}
		locator := ReportLocator{Type: Hostname, Value: host, Distinguisher: r.Distinguisher}
		if _, is4 := isValidIPv4(host); is4 {
			locator.Type = IPv4
		}
		downstreamLocators, err := locator.Implied()
		return append(ret, downstreamLocators...), err
	}
	return ret, nil
}

type ReportDistinguisher struct {
	Type  string
	Value string
}

type Finding struct {
	Identifier            string
	Name                  string
	OrganizationId        int
	ReportDistinguisher   ReportDistinguisher
	ReportLocator         ReportLocator
	ImpliedReportLocators []ReportLocator
}
