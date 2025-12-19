CREATE TABLE "users" (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
  email varchar(255) UNIQUE NOT NULL,
  name varchar(255) NOT NULL,
  dp TEXT, 
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_by uuid NOT NULL,
  updated_at timestamptz,
  updated_by uuid,
  deactivated_at timestamptz,
  deactivated_by uuid,
  deleted_at timestamptz,
  deleted_by uuid,
  PRIMARY KEY("id"),
  FOREIGN KEY("created_by") REFERENCES "users"("id"),
  FOREIGN KEY("updated_by") REFERENCES "users"("id"), 
  FOREIGN KEY("deleted_by") REFERENCES "users"("id")
);