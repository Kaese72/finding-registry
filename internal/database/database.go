package database

import "github.com/Kaese72/finding-registry/internal/intermediaries"

type Persistence interface {
	UpdateFinding(intermediaries.Finding, int) (intermediaries.Finding, error)
	GetFinding(string, int) (intermediaries.Finding, error)
	GetFindings(int) ([]intermediaries.Finding, error)
	Purge() error
}
