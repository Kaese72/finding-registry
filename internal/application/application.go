package application

import (
	"context"
	"errors"

	"github.com/Kaese72/finding-registry/event"
	"github.com/Kaese72/finding-registry/internal/database"
	"github.com/Kaese72/finding-registry/internal/intermediaries"
)

type ApplicationLogic struct {
	persistence    database.Persistence
	findingUpdates chan event.FindingUpdate
}

func NewApplicationLogic(persistence database.Persistence, findingUpdates chan event.FindingUpdate) ApplicationLogic {
	return ApplicationLogic{
		persistence:    persistence,
		findingUpdates: findingUpdates,
	}
}

func (logic ApplicationLogic) ReadFinding(ctx context.Context, identifier string, organizationID int) (intermediaries.Finding, error) {
	return logic.persistence.GetFinding(ctx, identifier, organizationID)
}

func (logic ApplicationLogic) ReadFindings(ctx context.Context, organizationID int) ([]intermediaries.Finding, error) {
	return logic.persistence.GetFindings(ctx, organizationID)
}

func (logic ApplicationLogic) PostFinding(ctx context.Context, finding intermediaries.Finding, organizationID int) (intermediaries.Finding, error) {
	finding.Identifier = "" // Do not allow identifier to be set
	if finding.ReportDistinguisher.Type == "" {
		return intermediaries.Finding{}, errors.New("must set report distingusher type")
	}
	if finding.ReportDistinguisher.Value == "" {
		return intermediaries.Finding{}, errors.New("must set report distingusher value")
	}
	if finding.ReportLocator.Type == "" {
		return intermediaries.Finding{}, errors.New("must set report locator type")
	}
	if finding.ReportLocator.Value == "" {
		return intermediaries.Finding{}, errors.New("must set report locator value")
	}
	if finding.ReportLocator.Distinguisher == "" {
		// If the locator is not set, we default to "global", indicating
		// it has no locality.
		finding.ReportLocator.Distinguisher = intermediaries.GlobalDistinguisher
	}
	implied, err := finding.ReportLocator.Implied()
	if err != nil {
		return intermediaries.Finding{}, err
	}
	finding.ImpliedReportLocators = implied
	resFinding, err := logic.persistence.UpdateFinding(ctx, finding, organizationID)
	if err != nil {
		return resFinding, err
	}
	logic.findingUpdates <- event.FindingUpdate{
		ID:             resFinding.Identifier,
		OrganizationId: resFinding.OrganizationId,
		ReportLocator: event.ReportLocator{
			Type:          string(resFinding.ReportLocator.Type),
			Value:         resFinding.ReportLocator.Value,
			Distinguisher: resFinding.ReportLocator.Distinguisher,
		},
	}
	return resFinding, err
}
