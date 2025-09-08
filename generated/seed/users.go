package seed

import (
	"context"

	"time"
	"github.com/jmoiron/sqlx"
)

type UsersRecordInput struct { 
  Id string
  CreatedAt time.Time
  UpdatedAt time.Time
  DeletedAt *time.Time
  IdpId string
  Username string
  Email string
  FullName string
  Appellation string
  AccountStatus string
}

type UsersRecord struct { 
  Id string
  CreatedAt time.Time
  UpdatedAt time.Time
  DeletedAt *time.Time
  IdpId string
  Username string
  Email string
  FullName string
  Appellation string
  AccountStatus string
}

func CreateUsersTableRecord(
  input UsersRecordInput,
) *UsersRecord {
  return &UsersRecord{ 
    AccountStatus: input.AccountStatus,
    Appellation: input.Appellation,
    CreatedAt: input.CreatedAt,
    DeletedAt: input.DeletedAt,
    Email: input.Email,
    FullName: input.FullName,
    Id: input.Id,
    IdpId: input.IdpId,
    UpdatedAt: input.UpdatedAt,
    Username: input.Username,
  }
}

func InsertUsersTableRecord(ctx context.Context, db *sqlx.DB, record UsersRecord) error {
  query := `
    INSERT INTO users (
      id,
      created_at,
      updated_at,
      deleted_at,
      idp_id,
      username,
      email,
      full_name,
      appellation,
      account_status
    )
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
  `
  _, err := db.ExecContext(ctx, query,
    record.Id,
    record.CreatedAt,
    record.UpdatedAt,
    record.DeletedAt,
    record.IdpId,
    record.Username,
    record.Email,
    record.FullName,
    record.Appellation,
    record.AccountStatus,
  )
  return err
}
