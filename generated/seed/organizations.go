package seed

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type OrganizationsRecordInput struct { 
  Id string
  CreatedAt string
  UpdatedAt string
  DeletedAt *string
  Name string
  Description string
}

type OrganizationsRecord struct { 
  Id string
  CreatedAt string
  UpdatedAt string
  DeletedAt *string
  Name string
  Description string
  AdminMemberId string
  BrokerMemberId string
}

func CreateOrganizationsTableRecord(
  input OrganizationsRecordInput,
  UsersModel UsersRecord,
) *OrganizationsRecord {
  return &OrganizationsRecord{ 
    AdminMemberId: UsersModel.Id,
    BrokerMemberId: UsersModel.Id,
    CreatedAt: input.CreatedAt,
    DeletedAt: input.DeletedAt,
    Description: input.Description,
    Id: input.Id,
    Name: input.Name,
    UpdatedAt: input.UpdatedAt,
  }
}

func InsertOrganizationsTableRecord(ctx context.Context, db *sqlx.DB, record OrganizationsRecord) error {
  query := `
    INSERT INTO organizations (
      id,
      created_at,
      updated_at,
      deleted_at,
      name,
      description,
      admin_member_id,
      broker_member_id
    )
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
  `
  _, err := db.ExecContext(ctx, query,
    record.Id,
    record.CreatedAt,
    record.UpdatedAt,
    record.DeletedAt,
    record.Name,
    record.Description,
    record.AdminMemberId,
    record.BrokerMemberId,
  )
  return err
}
