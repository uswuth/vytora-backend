package resolver

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/entity/category"
	complrec "github.com/uswuth/vytora-backend/internal/entity/compliance_record"
	"github.com/uswuth/vytora-backend/internal/entity/contract"
	"github.com/uswuth/vytora-backend/internal/entity/risk_assessment"
	"github.com/uswuth/vytora-backend/internal/entity/user"
	"github.com/uswuth/vytora-backend/internal/entity/vendor"
	"github.com/uswuth/vytora-backend/internal/graphql/model"
	"github.com/uswuth/vytora-backend/internal/services"
	graphqlmiddleware "github.com/uswuth/vytora-backend/internal/middleware/graphql"
)

func GetClaims(ctx context.Context) (*services.Claims, error) {
	c, ok := ctx.Value(graphqlmiddleware.UserContextKey).(*services.Claims)
	if !ok || c == nil {
		return nil, fmt.Errorf("unauthorized")
	}
	return c, nil
}

func stringPtr(s string) *string {
	return &s
}

func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func uuidPtrToStrPtr(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}

func ptrTime(t *time.Time) *time.Time {
	return t
}

func ptrFloat64(f *float64) *float64 {
	return f
}

func mapCategoryToGraphQL(c *category.Category) *model.Category {
	return &model.Category{
		ID:          c.ID.String(),
		Code:        c.Code,
		Name:        c.Name,
		DisplayName: c.DisplayName,
		Description: strPtrIfNotEmpty(c.Description),
		Status:      model.CategoryStatus(c.Status),
		CreatedBy:   uuidPtrToStrPtr(c.CreatedBy),
		UpdatedBy:   uuidPtrToStrPtr(c.UpdatedBy),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.CreatedAt,
	}
}

func mapComplianceRecordToGraphQL(c *complrec.ComplianceRecord) *model.ComplianceRecord {
	return &model.ComplianceRecord{
		ID:                c.ID.String(),
		Code:              c.Code,
		VendorID:          c.VendorID.String(),
		VendorCode:        strPtrIfNotEmpty(c.VendorCode),
		CertificationType: model.CertificationType(c.CertificationType),
		Status:            model.ComplianceStatus(c.Status),
		ValidFrom:         ptrTime(c.ValidFrom),
		ValidUntil:        ptrTime(c.ValidUntil),
		IssuedBy:          strPtrIfNotEmpty(c.IssuedBy),
		EvidenceURL:       strPtrIfNotEmpty(c.EvidenceURL),
		ReviewedBy:        uuidPtrToStrPtr(c.ReviewedBy),
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

func mapContractToGraphQL(c *contract.Contract) *model.Contract {
	return &model.Contract{
		ID:             c.ID.String(),
		Code:           c.Code,
		VendorID:       c.VendorID.String(),
		VendorCode:     strPtrIfNotEmpty(c.VendorCode),
		ContractNumber: c.ContractNumber,
		StartDate:      c.StartDate,
		EndDate:        c.EndDate,
		ContractValue:  ptrFloat64(c.ContractValue),
		RenewalStatus:  model.RenewalStatus(c.RenewalStatus),
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func mapVendorToGraphQL(v *vendor.Vendor) *model.Vendor {
	return &model.Vendor{
		ID:                    v.ID.String(),
		Code:                  v.Code,
		Name:                  v.Name,
		Category:              v.Category,
		ContactPerson:         strPtrIfNotEmpty(v.ContactPerson),
		ContactEmail:          strPtrIfNotEmpty(v.ContactEmail),
		Country:               strPtrIfNotEmpty(v.Country),
		RiskLevel:             model.RiskLevel(v.RiskLevel),
		Status:                model.VendorStatus(v.Status),
		AssignedDeptManagerID: uuidPtrToStrPtr(v.AssignedDeptManagerID),
		CreatedBy:             v.CreatedBy.String(),
		CreatedAt:             v.CreatedAt,
		UpdatedAt:             v.UpdatedAt,
	}
}

func mapUserToGraphQL(u *user.User) *model.User {
	return &model.User{
		ID:        u.ID.String(),
		Code:      u.Code,
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      model.UserRole(u.Role),
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func mapRiskAssessmentToGraphQL(r *risk_assessment.RiskAssessment) *model.RiskAssessment {
	return &model.RiskAssessment{
		ID:                   r.ID.String(),
		Code:                 r.Code,
		VendorID:             r.VendorID.String(),
		VendorCode:           strPtrIfNotEmpty(r.VendorCode),
		AssessmentDate:       r.AssessmentDate,
		AssessorID:           uuidPtrToStrPtr(r.AssessorID),
		OverallRiskScore:     r.OverallRiskScore,
		RiskLevel:            model.RiskLevel(r.RiskLevel),
		SecurityRiskScore:    r.SecurityRiskScore,
		FinancialRiskScore:   r.FinancialRiskScore,
		OperationalRiskScore: r.OperationalRiskScore,
		LegalRiskScore:       r.LegalRiskScore,
		Status:               model.AssessmentStatus(r.Status),
		Notes:                strPtrIfNotEmpty(r.Notes),
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}

// derefString safely dereferences a *string, returning empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func GetRawToken(ctx context.Context) (string, bool) {
	return graphqlmiddleware.GetRawToken(ctx)
}

func getDepartmentManagerID(ctx context.Context) *uuid.UUID {
	claims, err := GetClaims(ctx)
	if err != nil || claims.Role != "department_manager" {
		return nil
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil
	}
	return &uid
}
