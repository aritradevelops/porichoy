package db

import (
	"fmt"

	"github.com/aritradeveops/porichoy/internal/persistence/repository"
)

type Database interface {
	Connect() error
	Disconnect() error
	Health() error
	Tx() (repository.DBTX, error)
}

func NotInitializedErr(what string) error {
	return fmt.Errorf("%s is not initialized, have you forgot to call Connect() ?", what)
}
