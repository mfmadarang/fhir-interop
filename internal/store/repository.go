package store

import (
	"fmt"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// persists every Patient, Encounter, and Observation in a ParsedBundle. Existing records with the same id are updated
func SaveBundle(db *gorm.DB, pb *fhir.ParsedBundle) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, p := range pb.Patients {
			rec, err := patientToRecord(p)
			if err != nil {
				return err
			}
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&rec).Error; err != nil {
				return fmt.Errorf("saving patient %s: %w", p.ID, err)
			}
		}
		for _, e := range pb.Encounters {
			rec, err := encounterToRecord(e)
			if err != nil {
				return err
			}
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&rec).Error; err != nil {
				return fmt.Errorf("saving encounter %s: %w", e.ID, err)
			}
		}
		for _, o := range pb.Observations {
			rec, err := observationToRecord(o)
			if err != nil {
				return err
			}
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&rec).Error; err != nil {
				return fmt.Errorf("saving observation %s: %w", o.ID, err)
			}
		}
		return nil
	})
}