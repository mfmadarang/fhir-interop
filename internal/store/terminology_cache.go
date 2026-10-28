package store

import (
	"time"

	"github.com/mfmadarang/fhir-interop/internal/terminology"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// one answer from tx.fhir.org for a system and code
type TerminologyCacheRecord struct {
	System    string `gorm:"primaryKey"`
	Code      string `gorm:"primaryKey"`
	Valid     bool
	Message   string
	Display   string
	CheckedAt time.Time
}

func (TerminologyCacheRecord) TableName() string { return "terminology_cache" }

type TerminologyCache struct {
	db *gorm.DB
}

func NewTerminologyCache(db *gorm.DB) *TerminologyCache {
	return &TerminologyCache{db: db}
}

func (t *TerminologyCache) GetCached(system, code string) (*terminology.Result, time.Time, bool, error) {
	// Find with Limit, so a miss is not logged as an error like First would
	var recs []TerminologyCacheRecord
	err := t.db.Where("system = ? AND code = ?", system, code).Limit(1).Find(&recs).Error
	if err != nil {
		return nil, time.Time{}, false, err
	}
	if len(recs) == 0 {
		return nil, time.Time{}, false, nil
	}
	rec := recs[0]
	res := &terminology.Result{Valid: rec.Valid, Message: rec.Message, Display: rec.Display}
	return res, rec.CheckedAt, true, nil
}

func (t *TerminologyCache) PutCached(system, code string, res *terminology.Result) error {
	rec := TerminologyCacheRecord{
		System:    system,
		Code:      code,
		Valid:     res.Valid,
		Message:   res.Message,
		Display:   res.Display,
		CheckedAt: time.Now(),
	}
	return t.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&rec).Error
}
