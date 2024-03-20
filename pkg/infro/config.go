package infro

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/infro-io/infro-core/internal/argocd"
	"github.com/infro-io/infro-core/internal/github"
	"github.com/infro-io/infro-core/internal/model"
	"github.com/infro-io/infro-core/internal/terraform"
)

type (
	Config struct {
		Deployers []DeployerConfig `yaml:"deployers" validate:"required,dive"`
		VCS       VCSConfig        `yaml:"vcs" validate:"required"`
	}
	DeployerConfig struct {
		Value any `yaml:"value" validate:"required"`
	}
	VCSConfig struct {
		Value any `yaml:"value" validate:"required"`
	}
	ArgoCDConfig    = argocd.Config
	TerraformConfig = terraform.Config
	GithubConfig    = github.Config
)

func (dc *DeployerConfig) UnmarshalYAML(unmarshal func(any) error) error {
	var typ struct {
		Type string
	}
	if err := unmarshal(&typ); err != nil {
		return err
	}
	switch typ.Type {
	case model.DeployerTypeArgoCD:
		dc.Value = new(ArgoCDConfig)
	case model.DeployerTypeTerraform:
		dc.Value = new(TerraformConfig)
	}
	if err := unmarshal(dc.Value); err != nil {
		return err
	}
	return validator.New().Struct(dc.Value)
}

func (vc *VCSConfig) UnmarshalYAML(unmarshal func(any) error) error {
	var typ struct {
		Type string
	}
	if err := unmarshal(&typ); err != nil {
		return err
	}
	switch typ.Type {
	case model.VCSTypeGithub:
		vc.Value = new(GithubConfig)
	default:
		return fmt.Errorf("unrecognized vcs: %s", typ.Type)
	}
	if err := unmarshal(vc.Value); err != nil {
		return err
	}
	return validator.New().Struct(vc.Value)
}
