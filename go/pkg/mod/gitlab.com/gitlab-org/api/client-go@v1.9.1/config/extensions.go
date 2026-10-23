package config

import (
	"fmt"

	"buf.build/go/protoyaml"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

type Extension[T any] struct {
	name    string
	cfg     *Config
	context string
	ext     *T
}

func NewExtension[T any](name string, cfg *Config) *Extension[T] {
	e := &Extension[T]{
		name: name,
		cfg:  cfg,
		ext:  new(T),
	}
	return e
}

func NewExtensionForContext[T any](name string, cfg *Config, context string) *Extension[T] {
	e := &Extension[T]{
		name:    name,
		cfg:     cfg,
		context: context,
		ext:     new(T),
	}
	return e
}

func (e *Extension[T]) Unmarshal() (*T, error) {
	var ext *structpb.Struct
	if e.context == "" {
		ext = e.cfg.config.Extensions[e.name]
	} else {
		c := e.cfg.Context(e.context)
		if c == nil {
			return nil, fmt.Errorf("failed to find context %q", e.context)
		}

		i := e.cfg.Instance(*c.Instance)
		if i == nil {
			return nil, fmt.Errorf("failed to find instance %q from context %q", *c.Instance, e.context)
		}

		ext = i.Extensions[e.name]
	}

	if ext == nil {
		ext = &structpb.Struct{}
	}

	b, err := protoyaml.Marshal(ext)
	if err != nil {
		return nil, fmt.Errorf("failed to temporarily marshal extension to YAML: %w", err)
	}

	if err := yaml.Unmarshal(b, e.ext); err != nil {
		return nil, fmt.Errorf("failed to temporarily unmarshal to extension: %w", err)
	}
	return e.ext, nil
}

func (e *Extension[T]) Marshal() error {
	b, err := yaml.Marshal(e.ext)
	if err != nil {
		return fmt.Errorf("failed to temporarily marshal extension to YAML: %w", err)
	}

	var s structpb.Struct
	if err := protoyaml.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("failed to temporarily unmarshal to extension: %w", err)
	}

	if e.context == "" {
		if e.cfg.config.Extensions == nil {
			e.cfg.config.Extensions = make(map[string]*structpb.Struct, 1)
		}
		e.cfg.config.Extensions[e.name] = &s
	} else {
		c := e.cfg.Context(e.context)
		if c == nil {
			return fmt.Errorf("failed to find context %q", e.context)
		}

		i := e.cfg.Instance(*c.Instance)
		if i == nil {
			return fmt.Errorf("failed to find instance %q from context %q", *c.Instance, e.context)
		}

		if i.Extensions == nil {
			i.Extensions = make(map[string]*structpb.Struct, 1)
		}
		i.Extensions[e.name] = &s
	}

	return nil
}
