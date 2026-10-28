package store

import (
	"time"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/validate"
	"gorm.io/gorm"
)

const (
	SourceStructural  = "structural"
	SourceTerminology = "terminology"
)

// one problem validation found on a saved resource
type ValidationIssueRecord struct {
	ID           uint   `gorm:"primaryKey"`
	ResourceType string `gorm:"index:idx_validation_resource"`
	ResourceID   string `gorm:"index:idx_validation_resource"`
	Source       string
	Field        string
	Message      string
	CreatedAt    time.Time
}

func (ValidationIssueRecord) TableName() string { return "validation_issues" }

// replaces the stored issues of one source for every resource in the bundle, so a
// resource that passes now loses the issues it had before
func SaveValidationIssues(db *gorm.DB, pb *fhir.ParsedBundle, source string, issues []validate.Issue) error {
	return db.Transaction(func(tx *gorm.DB) error {
		drop := func(resourceType, id string) error {
			return tx.Where("resource_type = ? AND resource_id = ? AND source = ?", resourceType, id, source).
				Delete(&ValidationIssueRecord{}).Error
		}
		for _, p := range pb.Patients {
			if err := drop("Patient", p.ID); err != nil {
				return err
			}
		}
		for _, e := range pb.Encounters {
			if err := drop("Encounter", e.ID); err != nil {
				return err
			}
		}
		for _, o := range pb.Observations {
			if err := drop("Observation", o.ID); err != nil {
				return err
			}
		}

		if len(issues) == 0 {
			return nil
		}
		recs := make([]ValidationIssueRecord, 0, len(issues))
		for _, iss := range issues {
			recs = append(recs, ValidationIssueRecord{
				ResourceType: iss.ResourceType,
				ResourceID:   iss.ResourceID,
				Source:       source,
				Field:        iss.Field,
				Message:      iss.Message,
			})
		}
		return tx.Create(&recs).Error
	})
}

func ListValidationIssues(db *gorm.DB, resourceType, id string) ([]ValidationIssueRecord, error) {
	var recs []ValidationIssueRecord
	err := db.Where("resource_type = ? AND resource_id = ?", resourceType, id).Order("id").Find(&recs).Error
	return recs, err
}
