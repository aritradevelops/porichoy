-- Create "oauth_calls" table
CREATE TABLE "public"."oauth_calls" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "app_id" uuid NOT NULL,
  "code" character varying(255) NOT NULL,
  "user_id" uuid NOT NULL,
  "expires_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "oauth_calls_app_id_fkey" FOREIGN KEY ("app_id") REFERENCES "public"."apps" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "oauth_calls_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
