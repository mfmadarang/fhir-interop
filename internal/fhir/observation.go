package fhir

// FHIR's value[x] is polymorphic (valueQuantity, valueString, valueCodeableConcept, etc)
type Observation struct {
	ResourceType         string            `json:"resourceType"`
	ID                   string            `json:"id"`
	Status               string            `json:"status,omitempty"`
	Category             []CodeableConcept `json:"category,omitempty"`
	Code                 CodeableConcept   `json:"code"`
	Subject              Reference         `json:"subject,omitempty"`
	Encounter            Reference         `json:"encounter,omitempty"`
	EffectiveDateTime    string            `json:"effectiveDateTime,omitempty"`
	ValueQuantity        *Quantity         `json:"valueQuantity,omitempty"`
	ValueString          *string           `json:"valueString,omitempty"`
	ValueCodeableConcept *CodeableConcept  `json:"valueCodeableConcept,omitempty"`
}
