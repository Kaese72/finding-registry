package intermediaries

import (
	"fmt"
	"net"
	"net/url"
)

type ReportLocatorType string

const (
	IPv4     ReportLocatorType = "IPv4"
	Hostname ReportLocatorType = "Hostname"
	HTTP     ReportLocatorType = "HTTP"
	TCP      ReportLocatorType = "TCP"
	UDP      ReportLocatorType = "UDP"
)

func isValidIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

func (locator ReportLocator) Validate() error {
	switch locator.Type {
	case IPv4:
		// Validate IPv4 address
		if !isValidIPv4(locator.Value) {
			return fmt.Errorf("invalid IPv4 address: %s", locator.Value)
		}
	case HTTP:
		// Validate URL
		_, err := url.Parse(locator.Value)
		if err != nil {
			return fmt.Errorf("invalid URL: %s", locator.Value)
		}
	case TCP:
		// Validate TCP address
		_, err := net.ResolveTCPAddr("tcp", locator.Value)
		if err != nil {
			return fmt.Errorf("invalid TCP address: %s", locator.Value)
		}
	case UDP:
		// Validate UDP address
		_, err := net.ResolveUDPAddr("udp", locator.Value)
		if err != nil {
			return fmt.Errorf("invalid UDP address: %s", locator.Value)
		}
	case Hostname:
		// Everything is a valid hostname it seems
	default:
		return fmt.Errorf("invalid ReportLocatorType: %s", locator.Type)
	}
	return nil
}

type ReportLocator struct {
	Type  ReportLocatorType
	Value string
}

func (r ReportLocator) Implied() ([]ReportLocator, error) {
	if r.Validate() != nil {
		return nil, fmt.Errorf("invalid ReportLocator: %s", r)
	}
	ret := []ReportLocator{r}
	switch r.Type {
	case HTTP:
		u, err := url.Parse(r.Value)
		if err != nil {
			// This should never happen, since we already validated the URL
			panic(err)
		}
		locator := ReportLocator{Type: TCP, Value: u.Host}
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
		locator := ReportLocator{Type: Hostname, Value: host}
		if isValidIPv4(host) {
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
