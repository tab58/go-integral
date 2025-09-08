package seed

import (
	"context"

	"time"
	"github.com/jmoiron/sqlx"
)

type OrganizationMembersRecordInput struct { 
  CreatedAt time.Time
  UpdatedAt time.Time
  Scopes []string
}

type OrganizationMembersRecord struct { 
  CreatedAt time.Time
  UpdatedAt time.Time
  UserId string
  OrgId string
  Scopes []string
}

func CreateOrganizationMembersTableRecord(
  input OrganizationMembersRecordInput,
  OrganizationsModel OrganizationsRecord,
  UsersModel UsersRecord,
) *OrganizationMembersRecord {
  return &OrganizationMembersRecord{ 
    CreatedAt: input.CreatedAt,
    OrgId: OrganizationsModel.Id,
    Scopes: input.Scopes,
    UpdatedAt: input.UpdatedAt,
    UserId: UsersModel.Id,
  }
}

func InsertOrganizationMembersTableRecord(ctx context.Context, db *sqlx.DB, record OrganizationMembersRecord) error {
  query := `
    INSERT INTO organization_members (
      created_at,
      updated_at,
      user_id,
      org_id,
      scopes
    )
    VALUES ($1,$2,$3,$4,$5)
  `
  _, err := db.ExecContext(ctx, query,
    record.CreatedAt,
    record.UpdatedAt,
    record.UserId,
    record.OrgId,
    record.Scopes,
  )
  return err
}
