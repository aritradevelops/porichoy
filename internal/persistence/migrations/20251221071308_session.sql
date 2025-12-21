-- Modify "oauth_configs" table
ALTER TABLE "public"."oauth_configs" ADD COLUMN "jwt_lifetime" character varying(10) NOT NULL, ADD COLUMN "refresh_token_lifetime" character varying(10) NOT NULL;
-- Create "session" table
CREATE TABLE "public"."session" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "app_id" uuid NOT NULL,
  "refresh_token" text NOT NULL,
  "user_ip" character varying(45) NOT NULL,
  "user_agent" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "created_by" uuid NOT NULL,
  "updated_at" timestamptz NULL,
  "updated_by" uuid NULL,
  "deleted_at" timestamptz NULL,
  "deleted_by" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "session_refresh_token_key" UNIQUE ("refresh_token"),
  CONSTRAINT "session_app_id_fkey" FOREIGN KEY ("app_id") REFERENCES "public"."apps" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "session_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "session_deleted_by_fkey" FOREIGN KEY ("deleted_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "session_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "session_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
