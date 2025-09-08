package seed

import (
	"context"

	"time"
	"github.com/jmoiron/sqlx"
)

type BusinessOrganizationMembersRecordInput struct { 
  CreatedAt time.Time
  UpdatedAt time.Time
  Scopes []string
}

type BusinessOrganizationMembersRecord struct { 
  CreatedAt time.Time
  UpdatedAt time.Time
  UserId string
  OrgId string
  BusinessId string
  Scopes []string
}

func CreateBusinessOrganizationMembersTableRecord(
  input BusinessOrganizationMembersRecordInput,
  BusinessesModel BusinessesRecord,
  OrganizationsModel OrganizationsRecord,
  UsersModel UsersRecord,
) *BusinessOrganizationMembersRecord {
  return &BusinessOrganizationMembersRecord{ 
    BusinessId: BusinessesModel.Id,
    CreatedAt: input.CreatedAt,
    OrgId: OrganizationsModel.Id,
    Scopes: input.Scopes,
    UpdatedAt: input.UpdatedAt,
    UserId: UsersModel.Id,
  }
}

func InsertBusinessOrganizationMembersTableRecord(ctx context.Context, db *sqlx.DB, record BusinessOrganizationMembersRecord) error {
  query := `
    INSERT INTO business_organization_members (
      created_at,
      updated_at,
      user_id,
      org_id,
      business_id,
      scopes
    )
    VALUES ($1,$2,$3,$4,$5,$6)
  `
  _, err := db.ExecContext(ctx, query,
    record.CreatedAt,
    record.UpdatedAt,
    record.UserId,
    record.OrgId,
    record.BusinessId,
    record.Scopes,
  )
  return err
}
