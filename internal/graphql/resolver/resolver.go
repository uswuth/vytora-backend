package resolver

import (
	"github.com/uswuth/vytora-backend/internal/entity/audit_trail"
	"github.com/uswuth/vytora-backend/internal/entity/category"
	"github.com/uswuth/vytora-backend/internal/entity/compliance_record"
	"github.com/uswuth/vytora-backend/internal/entity/contract"
	"github.com/uswuth/vytora-backend/internal/entity/report"
	"github.com/uswuth/vytora-backend/internal/entity/risk_assessment"
	"github.com/uswuth/vytora-backend/internal/entity/user"
	"github.com/uswuth/vytora-backend/internal/entity/vendor"
	"github.com/uswuth/vytora-backend/internal/entity/vendor_contact"
	"github.com/uswuth/vytora-backend/internal/services"
)

type Resolver struct {
	UserRepo            *user.Repository
	VendorRepo          *vendor.Repository
	RiskAssessmentRepo  *risk_assessment.Repository
	ComplianceRepo      *compliance_record.Repository
	ContractRepo        *contract.Repository
	AuditRepo           *audit_trail.Repository
	ReportRepo          *report.Repository
	CategoryRepo        *category.Repository
	ContactRepo         *vendor_contact.Repository
	JWTService          *services.JWTService
	SeqService          *services.SequenceService
	AuditLogger         *audit_trail.Logger
}