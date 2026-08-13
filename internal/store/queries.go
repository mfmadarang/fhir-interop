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

func ListPatients(db *gorm.DB, limit, offset int) ([]*PatientRecord, error) {
	var recs []*PatientRecord
	if err := db.Order("id").Limit(limit).Offset(offset).Find(&recs).Error; err != nil {
		return nil, err
	}
	return recs, nil
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
