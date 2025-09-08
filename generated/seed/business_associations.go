package seed

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type BusinessAssociationsRecordInput struct { 
  CreatedAt string
}

type BusinessAssociationsRecord struct { 
  CreatedAt string
  BusinessId string
  OrgId string
}

func CreateBusinessAssociationsTableRecord(
  input BusinessAssociationsRecordInput,
  BusinessesModel BusinessesRecord,
  OrganizationsModel OrganizationsRecord,
) *BusinessAssociationsRecord {
  return &BusinessAssociationsRecord{ 
    BusinessId: BusinessesModel.Id,
    CreatedAt: input.CreatedAt,
    OrgId: OrganizationsModel.Id,
  }
}

func InsertBusinessAssociationsTableRecord(ctx context.Context, db *sqlx.DB, record BusinessAssociationsRecord) error {
  query := `
    INSERT INTO business_associations (
      created_at,
      business_id,
      org_id
    )
    VALUES ($1,$2,$3)
  `
  _, err := db.ExecContext(ctx, query,
    record.CreatedAt,
    record.BusinessId,
    record.OrgId,
  )
  return err
}
