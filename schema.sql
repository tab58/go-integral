-- user service table
CREATE TABLE "users" (
  "id" text NOT NULL PRIMARY KEY,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz DEFAULT NULL,

  "idp_id" text NOT NULL UNIQUE,
  "username" text NOT NULL UNIQUE,
  "email" text NOT NULL UNIQUE,
  "full_name" text NOT NULL,
  "appellation" text NOT NULL,
  "account_status" text NOT NULL DEFAULT 'inactive'
);

-- auth service table
CREATE TABLE "organizations" (
  "id" text NOT NULL PRIMARY KEY,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz DEFAULT NULL,

  "name" text NOT NULL,
  "description" text NOT NULL DEFAULT '',
  "admin_member_id" text NOT NULL,
  "broker_member_id" text NOT NULL,

  FOREIGN KEY ("admin_member_id") REFERENCES "users" ("id"),
  FOREIGN KEY ("broker_member_id") REFERENCES "users" ("id")
);

-- business service tables
CREATE TABLE "businesses" (
  "id" text NOT NULL PRIMARY KEY,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz DEFAULT NULL,

  "ein" text NOT NULL,
  "business_name" text NOT NULL,
  "business_address" text NOT NULL,
  "industry" text NOT NULL,
  "dba_name" text DEFAULT NULL,
  "owning_organization_id" text NOT NULL,

  FOREIGN KEY ("owning_organization_id") REFERENCES "organizations" ("id")
);

-- auth service table
CREATE TABLE "organization_members" (
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  -- we will actually delete team members

  "user_id" text NOT NULL,
  "org_id" text NOT NULL,
  "scopes" text[] NOT NULL DEFAULT '{}', -- populated by auth service

  PRIMARY KEY ("user_id", "org_id"),
  FOREIGN KEY ("user_id") REFERENCES "users" ("id"),
  FOREIGN KEY ("org_id") REFERENCES "organizations" ("id")
);

CREATE INDEX "idx_organization_members_org_id" ON "organization_members" ("org_id");
CREATE INDEX "idx_organization_members_user_id" ON "organization_members" ("user_id");

-- auth service table, quick lookup table of user scopes for a business
CREATE TABLE "business_organization_members" (
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "user_id" text NOT NULL,
  "org_id" text NOT NULL, -- here to allow a check if the user is a member of the organization
  "business_id" text NOT NULL,
  "scopes" text[] NOT NULL DEFAULT '{}',

  PRIMARY KEY ("user_id", "org_id", "business_id"),
  FOREIGN KEY ("user_id") REFERENCES "users" ("id"),
  FOREIGN KEY ("org_id") REFERENCES "organizations" ("id"),
  FOREIGN KEY ("business_id") REFERENCES "businesses" ("id")
);

CREATE INDEX "idx_business_organization_members_org_id" ON "business_organization_members" ("org_id");
CREATE INDEX "idx_business_organization_members_business_id" ON "business_organization_members" ("business_id");
CREATE INDEX "idx_business_organization_member_scopes_user_id_business_id" ON "business_organization_members" ("user_id", "business_id");
CREATE INDEX "idx_business_organization_members_user_id" ON "business_organization_members" ("user_id", "org_id");

-- auth service table, controls which organizations are associated with which businesses
CREATE TABLE "business_associations" (
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "business_id" text NOT NULL UNIQUE,
  "org_id" text NOT NULL,

  PRIMARY KEY ("business_id", "org_id"),
  FOREIGN KEY ("business_id") REFERENCES "businesses" ("id"),
  FOREIGN KEY ("org_id") REFERENCES "organizations" ("id")
);

CREATE INDEX "idx_business_associations_org_id" ON "business_associations" ("org_id");

CREATE TABLE "resources" (
  "id" text NOT NULL PRIMARY KEY,
  "uploaded_at" timestamptz NOT NULL,
  "deleted_at" timestamptz DEFAULT NULL,
  
  "uploader_id" text NOT NULL,
  "upload_bucket" text NOT NULL,
  "upload_key" text NOT NULL,
  "resource_filetype" text NOT NULL,

  "resource_filename" text DEFAULT NULL,
  "resource_size" bigint DEFAULT NULL,
  "tags" text[] DEFAULT NULL
);

CREATE INDEX "idx_resources_upload_bucket_and_key" ON "resources" ("upload_bucket", "upload_key");