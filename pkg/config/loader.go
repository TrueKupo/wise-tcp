package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Loader[T any] interface {
	Load(path string) (*T, error)
}

type LoaderOption[T any] func(l *YamlLoader[T])

type EnvMapper func(v *viper.Viper) error

func WithEnvMapper[T any](envMapper EnvMapper) LoaderOption[T] {
	return func(l *YamlLoader[T]) {
		l.envMapper = envMapper
	}
}

type YamlLoader[T any] struct {
	envMapper EnvMapper
}

func NewYamlLoader[T any](opts ...LoaderOption[T]) Loader[T] {
	l := &YamlLoader[T]{}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *YamlLoader[T]) Load(path string) (*T, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	var config T

	if l.envMapper != nil {
		if err := l.envMapper(v); err != nil {
			return nil, fmt.Errorf("failed to apply environment mappings: %w", err)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := applyEnvOverrides(v, &config); err != nil {
		return nil, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	return &config, nil
}

func applyEnvOverrides[T any](v *viper.Viper, config *T) error {
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := bindEnvTags(v, config)
	if err != nil {
		return fmt.Errorf("failed to bind environment variables: %w", err)
	}

	if err = v.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func bindEnvTags(v *viper.Viper, config any) error {
	val := reflect.ValueOf(config)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	return processStruct(v, val, "")
}

func processStruct(v *viper.Viper, val reflect.Value, parent string) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.Struct {
			nestedParent := parent + fieldType.Name + "."
			if err := processStruct(v, field, nestedParent); err != nil {
				return err
			}
			continue
		}

		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}

		if envValue, exists := os.LookupEnv(envTag); exists {
			v.Set(strings.ToLower(envTag), envValue)
		}
	}

	return nil
}

func MustLoad[T any](path string, opts ...LoaderOption[T]) *T {
	loader := NewYamlLoader(opts...)
	config, err := loader.Load(path)
	if err != nil {
		log.Fatal(err)
	}
	return config
}
