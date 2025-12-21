CREATE TABLE "oauth_calls" (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
  app_id uuid NOT NULL,
  code varchar(255) NOT NULL,
  user_id uuid NOT NULL,
  expires_at timestamptz NOT NULL,
  PRIMARY KEY("id"),
  FOREIGN KEY("app_id") REFERENCES "apps"("id"),
  FOREIGN KEY("user_id") REFERENCES "users"("id")
);