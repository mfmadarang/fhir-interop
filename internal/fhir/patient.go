package fhir

// patient is a FHIR R4 patient resource
type Patient struct {
	ResourceType string `json:"resourceType"`
	ID string `json:"id"`
	Identifier []Identifier `json:"identifier,omitempty"`
	Name []HumanName `json:"name,omitempty"`
	Gender string `json:"gender`
	BirthDate string `json:"birthDate,omitempty"`
	Address []Address `json:"address,omitempty"`
	Telecom []ContactPoint `json:"telecom,omitempty"`
}