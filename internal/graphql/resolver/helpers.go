package resolver

import (
	"context"
	"fmt"

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

	"github.com/uswuth/vytora-backend/internal/common/ptr"
)

func GetClaims(ctx context.Context) (*services.Claims, error) {
	c, ok := ctx.Value(graphqlmiddleware.UserContextKey).(*services.Claims)
	if !ok || c == nil {
		return nil, fmt.Errorf("unauthorized")
	}
	return c, nil
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

// Mapper functions: convert entity types to GraphQL model types.

func mapCategoryToGraphQL(c *category.Category) *model.Category {
	return &model.Category{
		ID:          c.ID.String(),
		Code:        c.Code,
		Name:        c.Name,
		DisplayName: c.DisplayName,
		Description: ptr.StrIfNotEmpty(c.Description),
		Status:      model.CategoryStatus(c.Status),
		CreatedBy:   ptr.UUIDStr(c.CreatedBy),
		UpdatedBy:   ptr.UUIDStr(c.UpdatedBy),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.CreatedAt,
	}
}

func mapComplianceRecordToGraphQL(c *complrec.ComplianceRecord) *model.ComplianceRecord {
	return &model.ComplianceRecord{
		ID:                c.ID.String(),
		Code:              c.Code,
		VendorID:          c.VendorID.String(),
		VendorCode:        ptr.StrIfNotEmpty(c.VendorCode),
		CertificationType: model.CertificationType(c.CertificationType),
		Status:            model.ComplianceStatus(c.Status),
		ValidFrom:         c.ValidFrom,
		ValidUntil:        c.ValidUntil,
		IssuedBy:          ptr.StrIfNotEmpty(c.IssuedBy),
		EvidenceURL:       ptr.StrIfNotEmpty(c.EvidenceURL),
		ReviewedBy:        ptr.UUIDStr(c.ReviewedBy),
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

func mapContractToGraphQL(c *contract.Contract) *model.Contract {
	return &model.Contract{
		ID:             c.ID.String(),
		Code:           c.Code,
		VendorID:       c.VendorID.String(),
		VendorCode:     ptr.StrIfNotEmpty(c.VendorCode),
		ContractNumber: c.ContractNumber,
		StartDate:      c.StartDate,
		EndDate:        c.EndDate,
		ContractValue:  c.ContractValue,
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
		ContactPerson:         ptr.StrIfNotEmpty(v.ContactPerson),
		ContactEmail:          ptr.StrIfNotEmpty(v.ContactEmail),
		Country:               ptr.StrIfNotEmpty(v.Country),
		RiskLevel:             model.RiskLevel(v.RiskLevel),
		Status:                model.VendorStatus(v.Status),
		AssignedDeptManagerID: ptr.UUIDStr(v.AssignedDeptManagerID),
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
		VendorCode:           ptr.StrIfNotEmpty(r.VendorCode),
		AssessmentDate:       r.AssessmentDate,
		AssessorID:           ptr.UUIDStr(r.AssessorID),
		OverallRiskScore:     r.OverallRiskScore,
		RiskLevel:            model.RiskLevel(r.RiskLevel),
		SecurityRiskScore:    r.SecurityRiskScore,
		FinancialRiskScore:   r.FinancialRiskScore,
		OperationalRiskScore: r.OperationalRiskScore,
		LegalRiskScore:       r.LegalRiskScore,
		Status:               model.AssessmentStatus(r.Status),
		Notes:                ptr.StrIfNotEmpty(r.Notes),
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}