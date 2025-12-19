-- Modify "users" table
ALTER TABLE "public"."users" ADD COLUMN "dp" text NULL, ADD COLUMN "deactivated_at" timestamptz NULL, ADD COLUMN "deactivated_by" uuid NULL;
