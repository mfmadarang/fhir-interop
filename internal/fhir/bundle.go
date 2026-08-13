package fhir

import (
	"encoding/json"
	"fmt"
)

//  fhir r4 bundle resource. synthea exports one bundle
//  per patient, of type "transaction"

type Bundle struct {
	ResourceType string	   		`json:"resourceType"`
	Type		 string	   		`json:"type"`
	Entry		 []BundleEntry 	`json:"entry,omitempty"`
}

// bundle mixes many diff resource types in one array
type BundleEntry struct {
	FullUrl		string		`json:"fullUrl"`
	Resource	json.RawMessage	`json:"resource"`
}

type resourceTypeProbe struct {
	ResourceType string `json:"resourceType"`
}

type ParsedBundle struct {
	Patients 		[]Patient
	Encounters 		[]Encounter
	Observations 	[]Observation
	Other			[]string
}

func ParseBundle(data []byte) (*ParsedBundle, error) {
	var b Bundle
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("parsing bundle: %w", err)
	}
	if b.ResourceType != "Bundle" {
		return nil, fmt.Errorf("expected resourceType \"Bundle\", got %q", b.ResourceType)
	}

	parsed := &ParsedBundle{}
	for i, entry := range b.Entry {
		var probe resourceTypeProbe
		if err := json.Unmarshal(entry.Resource, &probe); err != nil {
			return nil, fmt.Errorf("entry %d: probing resourceType: %w", i, err)
		}

		switch probe.ResourceType {
			case "Patient":
				var p Patient
				if err := json.Unmarshal(entry.Resource, &p); err != nil {
					return nil, fmt.Errorf("entry %d: parsing Patient: %w", i, err)
				}
				parsed.Patients = append(parsed.Patients, p)
			case "Encounter":
				var e Encounter
				if err := json.Unmarshal(entry.Resource, &e); err != nil {
					return nil, fmt.Errorf("entry %d: parsing Encounter: %w", i, err)
				}
				parsed.Encounters = append(parsed.Encounters, e)
			case "Observation":
				var o Observation
				if err := json.Unmarshal(entry.Resource, &o); err != nil {
					return nil, fmt.Errorf("entry %d: parsing Observation: %w", i, err)
				}
				parsed.Observations = append(parsed.Observations, o)
			default:
				parsed.Other = append(parsed.Other, probe.ResourceType)
		}
	}

	return parsed, nil
}