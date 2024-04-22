package database

import (
	"context"

	"github.com/Kaese72/finding-registry/internal/intermediaries"
)

type Persistence interface {
	UpdateFinding(context.Context, intermediaries.Finding, int) (intermediaries.Finding, error)
	GetFinding(context.Context, string, int) (intermediaries.Finding, error)
	GetFindings(context.Context, int) ([]intermediaries.Finding, error)
}
