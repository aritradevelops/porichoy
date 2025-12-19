-- Create "apps" table
CREATE TABLE "public"."apps" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NOT NULL,
  "domain" character varying(255) NOT NULL,
  "landing_url" character varying(255) NOT NULL,
  "logo" text NULL,
  "client_id" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "created_by" uuid NOT NULL,
  "updated_at" timestamptz NULL,
  "updated_by" uuid NULL,
  "deactivated_at" timestamptz NULL,
  "deactivated_by" uuid NULL,
  "deleted_at" timestamptz NULL,
  "deleted_by" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "apps_client_id_key" UNIQUE ("client_id"),
  CONSTRAINT "apps_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "apps_deleted_by_fkey" FOREIGN KEY ("deleted_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "apps_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "oauth_configs" table
CREATE TABLE "public"."oauth_configs" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "client_secret" text NOT NULL,
  "redirect_uris" text[] NULL,
  "success_callback_url" text NOT NULL,
  "error_callback_url" text NOT NULL,
  "jwt_algo" character varying(10) NOT NULL,
  "jwt_secret_resolver" text NULL,
  "app_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "created_by" uuid NOT NULL,
  "updated_at" timestamptz NULL,
  "updated_by" uuid NULL,
  "deleted_at" timestamptz NULL,
  "deleted_by" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "oauth_configs_app_id_fkey" FOREIGN KEY ("app_id") REFERENCES "public"."apps" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "oauth_configs_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "oauth_configs_deleted_by_fkey" FOREIGN KEY ("deleted_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "oauth_configs_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
