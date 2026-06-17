package middleware

var RolePermissions = map[string][]string{
	"system_admin": {
		"canCreateVendor", "canEditVendor", "canDeleteVendor",
		"canManageUsers", "canManageRoles", "canAccessAllReports",
		"canConfigureSystem",
		"canCreateRiskAssessment", "canReviewRisk", "canApproveRisk",
		"canReviewCompliance", "canTrackCertifications",
		"canSubmitVendorRequest", "canViewAssignedVendors",
		"canViewAuditHistory", "canDownloadReports", "canManageCategories",
	},
	"risk_manager": {
		"canCreateRiskAssessment", "canReviewRisk", "canApproveRisk",
		"canGenerateReports","canViewCategories",
	},
	"compliance_officer": {
		"canReviewCompliance", "canTrackCertifications", "canVerifyRegulations","canViewCategories",
	},
	"department_manager": {
		"canSubmitVendorRequest", "canViewAssignedVendors", "canRequestVendorReview","canViewCategories",
	},
	"auditor": {
		"canViewAuditHistory", "canDownloadReports", "canAccessAllReports","canViewCategories",
	},
}
