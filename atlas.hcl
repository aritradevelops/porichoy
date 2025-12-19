env "local" {
  # The database that Atlas will diff against
  url = "postgresql://postgres:admin@localhost:5432/porichoy?sslmode=disable"

  # Folder where migrations will be generated
  migration {
    dir = "file://internal/persistence/migrations"
  }

  # The schema.sql file to compare against
  dev = "docker://postgres/15/dev" # ephemeral dev db atlas uses for analysis
}

