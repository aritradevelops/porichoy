package resolver

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

type Resolver interface {
	Resolve(string) (any, error)
}

// resolves from env
type EnvResolver struct {
}

func (r *EnvResolver) Resolve(key string) (any, error) {
	key = strings.TrimPrefix(key, "env://")
	val, ok := os.LookupEnv(key)
	if !ok {
		return nil, fmt.Errorf("%s is not found in env", key)
	}
	return val, nil
}

// resolves literally
type LiteralResolver struct {
}

func (r *LiteralResolver) Resolve(key string) (any, error) {
	val := strings.TrimPrefix(key, "literal://")
	return val, nil
}

// TODO: db and s3
type ResolverFactory struct {
	providers map[string]Resolver
}

func NewResolverFactory() *ResolverFactory {
	providers := map[string]Resolver{}

	providers["env"] = &EnvResolver{}
	providers["literal"] = &LiteralResolver{}
	return &ResolverFactory{
		providers: providers,
	}
}

func (r *ResolverFactory) Providers() []string {
	pros := []string{}
	for k := range r.providers {
		pros = append(pros, k)
	}
	return pros
}

func (r *ResolverFactory) Auto(key string) (Resolver, error) {
	parts := strings.Split(key, "://")
	if len(parts) != 2 {
		return nil, fmt.Errorf("resolver_factory: invalid key")
	}
	id := parts[0]

	if !slices.Contains(r.Providers(), id) {
		return nil, fmt.Errorf("resolver_factory: provider not implemented for id: %s", id)
	}
	return r.providers[id], nil
}
