package client

import (
	"errors"
	"fmt"
	"github.com/hashicorp/hcl/v2"
)

type BackendType string

const (
	LOCAL BackendType = "local"
	S3    BackendType = "s3"
)

type Backend interface {
	loadConfig(config *hcl.Attribute) error
	Init()
}

type LocalBackend struct {
	Path string
}

func (b *LocalBackend) Init() {
	fmt.Print("load config")
	fmt.Print(b.Path)
}

func (b *LocalBackend) loadConfig(config *hcl.Attribute) error {
	val, err := config.Expr.Value(nil)
	if err != nil {
		return err
	}
	configValueMap := val.AsValueMap()
	path := configValueMap["path"].AsString()
	b.Path = path
	return nil
}

type S3Backend struct {
	Bucket string
	Key    string
	Region string
}

func (b *S3Backend) Init() {
	fmt.Print("load config")
	fmt.Print(b.Bucket)
	fmt.Print(b.Key)
	fmt.Print(b.Region)
}

func (b *S3Backend) loadConfig(config *hcl.Attribute) error {
	val, err := config.Expr.Value(nil)
	if err != nil {
		return err
	}
	configValueMap := val.AsValueMap()
	b.Bucket = configValueMap["bucket"].AsString()
	b.Key = configValueMap["key"].AsString()
	b.Region = configValueMap["region"].AsString()
	return nil
}

func NewBackend(cfg *Config) (Backend, error) {
	switch cfg.Backend {
	case LOCAL:
		backend := LocalBackend{}
		err := backend.loadConfig(cfg.Config)
		if err != nil {
			return nil, err
		}
		return &backend, nil
	case S3:
		backend := LocalBackend{}
		err := backend.loadConfig(cfg.Config)
		if err != nil {
			return nil, err
		}
		return &backend, nil
	default:
		return nil, errors.New("Not supported backend")
	}
}
