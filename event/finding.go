package event

type ReportLocator struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type FindingUpdate struct {
	ID             string        `json:"id"`
	OrganizationId int           `json:"organizationId"`
	ReportLocator  ReportLocator `json:"reportLocator"`
}
