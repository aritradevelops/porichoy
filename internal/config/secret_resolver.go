package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
)

var secretPattern = regexp.MustCompile(`^\$\{([A-Z0-9_]+)\}$`)

func resolveSecrets(v any) error {
	return resolve(reflect.ValueOf(v))
}

func resolve(v reflect.Value) error {
	if v.Kind() == reflect.Pointer {
		return resolve(v.Elem())
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := resolve(v.Field(i)); err != nil {
				return err
			}
		}

	case reflect.String:
		str := v.String()
		matches := secretPattern.FindStringSubmatch(str)
		if len(matches) == 2 {
			env := matches[1]
			value, ok := os.LookupEnv(env)
			if !ok {
				return fmt.Errorf("missing required secret: %s", env)
			}
			v.SetString(value)
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := resolve(v.Index(i)); err != nil {
				return err
			}
		}

	case reflect.Map:
		for _, key := range v.MapKeys() {
			if err := resolve(v.MapIndex(key)); err != nil {
				return err
			}
		}
	}

	return nil
}
