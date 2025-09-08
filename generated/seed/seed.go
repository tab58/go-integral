package seed

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type SchemaModels struct { 
  UsersModels []UsersRecord
  ResourcesModels []ResourcesRecord
  OrganizationsModels []OrganizationsRecord
  OrganizationMembersModels []OrganizationMembersRecord
  BusinessesModels []BusinessesRecord
  BusinessOrganizationMembersModels []BusinessOrganizationMembersRecord
  BusinessAssociationsModels []BusinessAssociationsRecord
}

func SeedDatabase(ctx context.Context, db *sqlx.DB, models SchemaModels) error { 
  for _, record := range models.UsersModels {
    err := InsertUsersTableRecord(ctx, db, record)
    if err != nil {
      return err
    }
  }
  
  for _, record := range models.ResourcesModels {
    err := InsertResourcesTableRecord(ctx, db, record)
    if err != nil {
      return err
    }
  }
  
  for _, record := range models.OrganizationsModels {
    err := InsertOrganizationsTableRecord(ctx, db, record)
    if err != nil {
      return err
    }
  }
  
  for _, record := range models.OrganizationMembersModels {
    err := InsertOrganizationMembersTableRecord(ctx, db, record)
    if err != nil {
      return err
    }
  }
  
  for _, record := range models.BusinessesModels {
    err := InsertBusinessesTableRecord(ctx, db, record)
    if err != nil {
      return err
    }
  }
  
  for _, record := range models.BusinessOrganizationMembersModels {
    err := InsertBusinessOrganizationMembersTableRecord(ctx, db, record)
    if err != nil {
      return err
    }
  }
  
  for _, record := range models.BusinessAssociationsModels {
    err := InsertBusinessAssociationsTableRecord(ctx, db, record)
    if err != nil {
      return err
    }
  }
  
  return nil
}