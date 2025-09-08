package seed

import (
	"context"

	"time"
	"github.com/jmoiron/sqlx"
)

type BusinessesRecordInput struct { 
  Id string
  CreatedAt time.Time
  UpdatedAt time.Time
  DeletedAt *time.Time
  Ein string
  BusinessName string
  BusinessAddress string
  Industry string
  DbaName *string
}

type BusinessesRecord struct { 
  Id string
  CreatedAt time.Time
  UpdatedAt time.Time
  DeletedAt *time.Time
  Ein string
  BusinessName string
  BusinessAddress string
  Industry string
  DbaName *string
  OwningOrganizationId string
}

func CreateBusinessesTableRecord(
  input BusinessesRecordInput,
  OrganizationsModel OrganizationsRecord,
) *BusinessesRecord {
  return &BusinessesRecord{ 
    BusinessAddress: input.BusinessAddress,
    BusinessName: input.BusinessName,
    CreatedAt: input.CreatedAt,
    DbaName: input.DbaName,
    DeletedAt: input.DeletedAt,
    Ein: input.Ein,
    Id: input.Id,
    Industry: input.Industry,
    OwningOrganizationId: OrganizationsModel.Id,
    UpdatedAt: input.UpdatedAt,
  }
}

func InsertBusinessesTableRecord(ctx context.Context, db *sqlx.DB, record BusinessesRecord) error {
  query := `
    INSERT INTO businesses (
      id,
      created_at,
      updated_at,
      deleted_at,
      ein,
      business_name,
      business_address,
      industry,
      dba_name,
      owning_organization_id
    )
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
  `
  _, err := db.ExecContext(ctx, query,
    record.Id,
    record.CreatedAt,
    record.UpdatedAt,
    record.DeletedAt,
    record.Ein,
    record.BusinessName,
    record.BusinessAddress,
    record.Industry,
    record.DbaName,
    record.OwningOrganizationId,
  )
  return err
}
