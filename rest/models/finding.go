package models

import "github.com/Kaese72/finding-registry/internal/intermediaries"

type ReportLocator struct {
	Type          string `json:"type"`
	Value         string `json:"value"`
	Distinguisher string `json:"distinguisher"`
}

func (locator ReportLocator) toIntermediary() intermediaries.ReportLocator {
	return intermediaries.ReportLocator{
		Type:          intermediaries.ReportLocatorType(locator.Type),
		Value:         locator.Value,
		Distinguisher: locator.Distinguisher,
	}
}

func ReportLocatorFromIntermediary(intermediary intermediaries.ReportLocator) ReportLocator {
	return ReportLocator{
		Type:          string(intermediary.Type),
		Value:         intermediary.Value,
		Distinguisher: intermediary.Distinguisher,
	}
}

type ReportDistinguisher struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (locator ReportDistinguisher) toIntermediary() intermediaries.ReportDistinguisher {
	return intermediaries.ReportDistinguisher{
		Type:  locator.Type,
		Value: locator.Value,
	}
}

func ReportDistinguisherFromIntermediary(intermediary intermediaries.ReportDistinguisher) ReportDistinguisher {
	return ReportDistinguisher{
		Type:  intermediary.Type,
		Value: intermediary.Value,
	}
}

type Finding struct {
	Identifier            string              `json:"identifier"`
	Name                  string              `json:"name"`
	OrganizationId        int                 `json:"organizationId"`
	ReportDistinguisher   ReportDistinguisher `json:"reportDistinguisher"`
	ReportLocator         ReportLocator       `json:"reportLocator"`
	ImpliedReportLocators []ReportLocator     `json:"impliedReportLocators"`
}

func (finding Finding) ToIntermediary() intermediaries.Finding {
	implied := []intermediaries.ReportLocator{}
	for index := range finding.ImpliedReportLocators {
		implied = append(implied, finding.ImpliedReportLocators[index].toIntermediary())
	}
	return intermediaries.Finding{
		Identifier:            finding.Identifier,
		Name:                  finding.Name,
		OrganizationId:        finding.OrganizationId,
		ReportDistinguisher:   finding.ReportDistinguisher.toIntermediary(),
		ReportLocator:         finding.ReportLocator.toIntermediary(),
		ImpliedReportLocators: implied,
	}
}

func FindingFromIntermediary(intermediary intermediaries.Finding) Finding {
	reportLocators := []ReportLocator{}
	for index := range intermediary.ImpliedReportLocators {
		reportLocators = append(reportLocators, ReportLocatorFromIntermediary(intermediary.ImpliedReportLocators[index]))
	}
	return Finding{
		Identifier:            intermediary.Identifier,
		Name:                  intermediary.Name,
		OrganizationId:        intermediary.OrganizationId,
		ReportDistinguisher:   ReportDistinguisherFromIntermediary(intermediary.ReportDistinguisher),
		ReportLocator:         ReportLocatorFromIntermediary(intermediary.ReportLocator),
		ImpliedReportLocators: reportLocators,
	}
}
