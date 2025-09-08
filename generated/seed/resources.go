package seed

import (
	"context"

	"time"
	"github.com/jmoiron/sqlx"
)

type ResourcesRecordInput struct { 
  Id string
  UploadedAt time.Time
  DeletedAt *time.Time
  UploaderId string
  UploadBucket string
  UploadKey string
  ResourceFiletype string
  ResourceFilename *string
  ResourceSize *string
  Tags *string
}

type ResourcesRecord struct { 
  Id string
  UploadedAt time.Time
  DeletedAt *time.Time
  UploaderId string
  UploadBucket string
  UploadKey string
  ResourceFiletype string
  ResourceFilename *string
  ResourceSize *string
  Tags *string
}

func CreateResourcesTableRecord(
  input ResourcesRecordInput,
) *ResourcesRecord {
  return &ResourcesRecord{ 
    DeletedAt: input.DeletedAt,
    Id: input.Id,
    ResourceFilename: input.ResourceFilename,
    ResourceFiletype: input.ResourceFiletype,
    ResourceSize: input.ResourceSize,
    Tags: input.Tags,
    UploadBucket: input.UploadBucket,
    UploadKey: input.UploadKey,
    UploadedAt: input.UploadedAt,
    UploaderId: input.UploaderId,
  }
}

func InsertResourcesTableRecord(ctx context.Context, db *sqlx.DB, record ResourcesRecord) error {
  query := `
    INSERT INTO resources (
      id,
      uploaded_at,
      deleted_at,
      uploader_id,
      upload_bucket,
      upload_key,
      resource_filetype,
      resource_filename,
      resource_size,
      tags
    )
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
  `
  _, err := db.ExecContext(ctx, query,
    record.Id,
    record.UploadedAt,
    record.DeletedAt,
    record.UploaderId,
    record.UploadBucket,
    record.UploadKey,
    record.ResourceFiletype,
    record.ResourceFilename,
    record.ResourceSize,
    record.Tags,
  )
  return err
}
