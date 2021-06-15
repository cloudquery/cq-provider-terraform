package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type BackendConfigBlock struct {
	BackendName string      `hcl:"config,label"`
	BackendType string      `hcl:"backend,attr"`
	ConfigAttrs interface{} `hcl:"config,remain"`
}

type BackendType string

const (
	LOCAL BackendType = "local"
	S3    BackendType = "s3"
)

type Backend interface {
	loadConfig(config interface{}) error
	read() (*TerraformData, error)
	Type() BackendType
	Name() string
	Data() *TerraformData
}

type LocalBackendConfig struct {
	Path string `hcl:"path"`
}

type LocalBackend struct {
	BackendName string
	Config      LocalBackendConfig
	data        *TerraformData
}

func (b *LocalBackend) read() (*TerraformData, error) {
	f, err := os.Open(b.Config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tfstate from %s", b.Config.Path)
	}
	defer f.Close()

	var s TerraformData
	if err := json.NewDecoder(f).Decode(&s.State); err != nil {
		return nil, fmt.Errorf("invalid tf state file")
	}
	if s.State.Version != StateVersion {
		return nil, fmt.Errorf("unsupported state version %d", s.State.Version)
	}
	return &s, nil
}

func (b *LocalBackend) loadConfig(config interface{}) error {
	cfg := config.(hcl.Body)
	if diags := gohcl.DecodeBody(cfg, nil, &b.Config); diags != nil {
		return errors.New("cannot parse backend config")
	}
	return nil
}

func (b *LocalBackend) Type() BackendType {
	return LOCAL
}

func (b *LocalBackend) Name() string {
	return b.BackendName
}

func (b *LocalBackend) Data() *TerraformData {
	return b.data
}

type S3BackendConfig struct {
	Bucket  string `hcl:"bucket"`
	Key     string `hcl:"key"`
	Region  string `hcl:"region"`
	RoleArn string `hcl:"role_arn,optional"`
}

type S3Backend struct {
	BackendName string
	Config      S3BackendConfig
	data        *TerraformData
}

func (b *S3Backend) read() (*TerraformData, error) {
	if b.Config.Region == "" {
		if region, err := s3manager.GetBucketRegion(
			context.Background(),
			session.Must(session.NewSession()),
			b.Config.Bucket,
			"us-east-1",
		); err != nil {
			return nil, err
		} else {
			b.Config.Region = region
		}
	}

	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(b.Config.Region),
		},
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return nil, err
	}

	cfg := &aws.Config{}
	if b.Config.RoleArn != "" {
		arn, err := arn.Parse(b.Config.RoleArn)
		if err != nil {
			return nil, err
		}
		creds := stscreds.NewCredentials(sess, arn.String())
		cfg.Credentials = creds
	}
	svc := s3.New(sess, cfg)

	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b.Config.Bucket),
		Key:    aws.String(b.Config.Key),
	})
	if err != nil {
		return nil, err
	}
	var s TerraformData
	if err := json.NewDecoder(result.Body).Decode(&s.State); err != nil {
		return nil, fmt.Errorf("invalid tf state file")
	}
	if s.State.Version != StateVersion {
		return nil, fmt.Errorf("unsupported state version %d", s.State.Version)
	}
	return &s, nil
}

func (b *S3Backend) loadConfig(config interface{}) error {
	cfg := config.(hcl.Body)
	if diags := gohcl.DecodeBody(cfg, nil, &b.Config); diags != nil {
		return errors.New("cannot parse backend config")
	}
	return nil
}

func (b *S3Backend) Type() BackendType {
	return S3
}

func (b *S3Backend) Name() string {
	return b.BackendName
}

func (b *S3Backend) Data() *TerraformData {
	return b.data
}

func NewBackend(cfg *BackendConfigBlock) (Backend, error) {
	switch cfg.BackendType {
	case "local":
		backend := LocalBackend{}
		backend.BackendName = cfg.BackendName
		if err := backend.loadConfig(cfg.ConfigAttrs); err != nil {
			return nil, err
		}
		if d, err := backend.read(); err != nil {
			return nil, err
		} else {
			backend.data = d
		}
		return &backend, nil
	case "s3":
		backend := S3Backend{}
		backend.BackendName = cfg.BackendName
		if err := backend.loadConfig(cfg.ConfigAttrs); err != nil {
			return nil, err
		}
		if d, err := backend.read(); err != nil {
			return nil, err
		} else {
			backend.data = d
		}
		return &backend, nil
	default:
		return nil, errors.New("unsupported backend")
	}
}
