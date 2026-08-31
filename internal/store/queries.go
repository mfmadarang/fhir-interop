package store

import "gorm.io/gorm"

func GetPatient(db *gorm.DB, id string) (*PatientRecord, error) {
	var rec PatientRecord
	err := db.First(&rec, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// PatientSearch holds the optional filters for SearchPatients. A zero value
// matches every patient.
type PatientSearch struct {
	Family    string
	Given     string
	Gender    string
	BirthDate string
	Limit     int
}

// searches patients by the set fields. Family and Given are case-insensitive
// prefix matches (FHIR string search semantics); Gender and BirthDate are exact.
func SearchPatients(db *gorm.DB, s PatientSearch) ([]*PatientRecord, error) {
	q := db.Order("id")
	if s.Family != "" {
		q = q.Where("family_name ILIKE ?", s.Family+"%")
	}
	if s.Given != "" {
		q = q.Where("given_name ILIKE ?", s.Given+"%")
	}
	if s.Gender != "" {
		q = q.Where("gender = ?", s.Gender)
	}
	if s.BirthDate != "" {
		q = q.Where("birth_date = ?", s.BirthDate)
	}
	if s.Limit > 0 {
		q = q.Limit(s.Limit)
	}

	var recs []*PatientRecord
	if err := q.Find(&recs).Error; err != nil {
		return nil, err
	}
	return recs, nil
}

func ListPatients(db *gorm.DB, limit, offset int) ([]*PatientRecord, error) {
	var recs []*PatientRecord
	if err := db.Order("id").Limit(limit).Offset(offset).Find(&recs).Error; err != nil {
		return nil, err
	}
	return recs, nil
}

// lists patients ordered by ID, cursor-paginated. empty afterID starts from the beginning
func ListPatientsCursor(db *gorm.DB, first int, afterID string) ([]*PatientRecord, bool, error) {
	q := db.Order("id").Limit(first + 1)
	if afterID != "" {
		q = q.Where("id > ?", afterID)
	}
	var recs []*PatientRecord
	if err := q.Find(&recs).Error; err != nil {
		return nil, false, err
	}
	hasNextPage := len(recs) > first
	if hasNextPage {
		recs = recs[:first]
	}
	return recs, hasNextPage, nil
}

func ListEncountersByPatient(db *gorm.DB, patientID string) ([]*EncounterRecord, error) {
	var recs []*EncounterRecord
	if err := db.Where("patient_id = ?", patientID).Order("period_start desc").Find(&recs).Error; err != nil {
		return nil, err
	}
	return recs, nil
}

func ListObservationsByPatient(db *gorm.DB, patientID string) ([]*ObservationRecord, error) {
	var recs []*ObservationRecord
	if err := db.Where("patient_id = ?", patientID).Order("effective_date_time desc").Find(&recs).Error; err != nil {
		return nil, err
	}
	return recs, nil
}
