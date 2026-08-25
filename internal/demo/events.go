package demo

type StageName string

const (
	StageParse       StageName = "parse"
	StageConvert     StageName = "convert"
	StageValidate    StageName = "validate"
	StageTerminology StageName = "terminology"
	StagePersist     StageName = "persist"
)

type StageStatus string

const (
	StatusRunning StageStatus = "running"
	StatusDone    StageStatus = "done"
	StatusError   StageStatus = "error"
)

type StageEvent struct {
	Stage   StageName   `json:"stage"`
	Status  StageStatus `json:"status"`
	Message string      `json:"message,omitempty"`
}

type TerminologyResult string

const (
	TerminologyValid      TerminologyResult = "valid"
	TerminologyInvalid    TerminologyResult = "invalid"
	TerminologyUnverified TerminologyResult = "unverified"
)

type TerminologyEvent struct {
	ObservationID string            `json:"observationId"`
	System        string            `json:"system"`
	Code          string            `json:"code"`
	Display       string            `json:"display,omitempty"`
	Result        TerminologyResult `json:"result"`
	Message       string            `json:"message,omitempty"`
	Simulated     bool              `json:"simulated,omitempty"`
}
