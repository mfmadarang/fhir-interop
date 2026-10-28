package fhir

import (
	"encoding/json"
	"fmt"
)

//  fhir r4 bundle resource. synthea exports one bundle
//  per patient, of type "transaction"

type Bundle struct {
	ResourceType string        `json:"resourceType"`
	Type         string        `json:"type"`
	Entry        []BundleEntry `json:"entry,omitempty"`
}

// bundle mixes many diff resource types in one array
type BundleEntry struct {
	FullUrl  string          `json:"fullUrl"`
	Resource json.RawMessage `json:"resource"`
}

type resourceTypeProbe struct {
	ResourceType string `json:"resourceType"`
}

type ParsedBundle struct {
	Patients     []Patient
	Encounters   []Encounter
	Observations []Observation
	Other        []string
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

		if err := parsed.add(probe.ResourceType, entry.Resource); err != nil {
			return nil, fmt.Errorf("entry %d: %w", i, err)
		}
	}

	return parsed, nil
}

// adds one resource to the right slice, based on its resourceType
func (pb *ParsedBundle) add(resourceType string, raw []byte) error {
	switch resourceType {
	case "Patient":
		var p Patient
		if err := json.Unmarshal(raw, &p); err != nil {
			return fmt.Errorf("parsing Patient: %w", err)
		}
		pb.Patients = append(pb.Patients, p)
	case "Encounter":
		var e Encounter
		if err := json.Unmarshal(raw, &e); err != nil {
			return fmt.Errorf("parsing Encounter: %w", err)
		}
		pb.Encounters = append(pb.Encounters, e)
	case "Observation":
		var o Observation
		if err := json.Unmarshal(raw, &o); err != nil {
			return fmt.Errorf("parsing Observation: %w", err)
		}
		pb.Observations = append(pb.Observations, o)
	default:
		pb.Other = append(pb.Other, resourceType)
	}
	return nil
}

// like ParseBundle, but also accepts a single Patient, Encounter or Observation
func ParseResource(data []byte) (*ParsedBundle, error) {
	var probe resourceTypeProbe
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("reading resourceType: %w", err)
	}

	switch probe.ResourceType {
	case "Bundle":
		return ParseBundle(data)
	case "Patient", "Encounter", "Observation":
		parsed := &ParsedBundle{}
		if err := parsed.add(probe.ResourceType, data); err != nil {
			return nil, err
		}
		return parsed, nil
	case "":
		return nil, fmt.Errorf("missing resourceType")
	default:
		return nil, fmt.Errorf("resourceType %q is not supported, send a Bundle, Patient, Encounter or Observation", probe.ResourceType)
	}
}
