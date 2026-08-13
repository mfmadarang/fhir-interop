package fhir

type Encounter struct {
	ResourceType string            `json:"resourceType"`
	ID           string            `json:"id"`
	Status       string            `json:"status,omitempty"`
	Class        Coding            `json:"class,omitempty"`
	Type         []CodeableConcept `json:"type,omitempty"`
	Subject      Reference         `json:"subject,omitempty"`
	Period       Period            `json:"period,omitempty"`
}
