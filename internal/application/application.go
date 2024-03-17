package application

import (
	"errors"

	"github.com/Kaese72/findings-registry/event"
	"github.com/Kaese72/findings-registry/internal/database"
	"github.com/Kaese72/findings-registry/internal/intermediaries"
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

func (logic ApplicationLogic) ReadFinding(identifier string) (intermediaries.Finding, error) {
	return logic.persistence.GetFinding(identifier)
}

func (logic ApplicationLogic) ReadFindings() ([]intermediaries.Finding, error) {
	return logic.persistence.GetFindings()
}

func (logic ApplicationLogic) PostFinding(finding intermediaries.Finding) (intermediaries.Finding, error) {
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
	implied, err := finding.ReportLocator.Implied()
	if err != nil {
		return intermediaries.Finding{}, err
	}
	finding.ImpliedReportLocators = implied
	resFinding, err := logic.persistence.UpdateFinding(finding)
	if err != nil {
		return resFinding, err
	}
	logic.findingUpdates <- event.FindingUpdate{
		ID:             resFinding.Identifier,
		OrganizationId: resFinding.OrganizationId,
		ReportLocator:  event.ReportLocator{Type: string(resFinding.ReportLocator.Type), Value: resFinding.ReportLocator.Value},
	}
	return resFinding, err
}
