package middleware

var RolePermissions = map[string][]string{
	"system_admin": {
		"canCreateVendor", "canEditVendor", "canDeleteVendor",
		"canManageUsers", "canManageRoles", "canAccessAllReports",
		"canConfigureSystem",
		"canCreateRiskAssessment", "canReviewRisk", "canApproveRisk",
		"canReviewCompliance", "canTrackCertifications",
		"canSubmitVendorRequest", "canViewAssignedVendors",
		"canViewAuditHistory", "canDownloadReports",
	},
	"risk_manager": {
		"canCreateRiskAssessment", "canReviewRisk", "canApproveRisk",
		"canGenerateReports",
	},
	"compliance_officer": {
		"canReviewCompliance", "canTrackCertifications", "canVerifyRegulations",
	},
	"department_manager": {
		"canSubmitVendorRequest", "canViewAssignedVendors", "canRequestVendorReview",
	},
	"auditor": {
		"canViewAuditHistory", "canDownloadReports", "canAccessAllReports",
	},
}
