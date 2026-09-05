package aifeedback

const (
	ReportReasonAlreadyCorrect           = "already_correct"
	ReportReasonCorrectionChangedMeaning = "correction_changed_meaning"
	ReportReasonExplanationUnclear       = "explanation_unclear"
	ReportReasonInappropriate            = "inappropriate"
	ReportReasonSomethingElse            = "something_else"
	QualityReviewStateOpen               = "open"
)

func validReportReason(reason string) bool {
	switch reason {
	case ReportReasonAlreadyCorrect, ReportReasonCorrectionChangedMeaning,
		ReportReasonExplanationUnclear, ReportReasonInappropriate,
		ReportReasonSomethingElse:
		return true
	default:
		return false
	}
}
