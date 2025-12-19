-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" character varying(255) NOT NULL,
  "name" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "created_by" uuid NOT NULL,
  "updated_at" timestamptz NULL,
  "updated_by" uuid NULL,
  "deleted_at" timestamptz NULL,
  "deleted_by" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email"),
  CONSTRAINT "users_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "users_deleted_by_fkey" FOREIGN KEY ("deleted_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "users_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "passwords" table
CREATE TABLE "public"."passwords" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "hashed_password" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "created_by" uuid NOT NULL,
  "updated_at" timestamptz NULL,
  "updated_by" uuid NULL,
  "deleted_at" timestamptz NULL,
  "deleted_by" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "passwords_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "passwords_deleted_by_fkey" FOREIGN KEY ("deleted_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "passwords_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
