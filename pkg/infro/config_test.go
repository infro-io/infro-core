package infro_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"infro.io/infro-core/pkg/infro"
)

const yamlConfig = `
deployers:
  - type: argocd
    name: example
    endpoint: http://example.org
    authtoken: token
vcs:
  type: github
  authtoken: token
`

func TestUnmarshalYAML(t *testing.T) {
	var cfg infro.Config
	err := yaml.Unmarshal([]byte(yamlConfig), &cfg)
	require.NoError(t, err)
	require.NotEmpty(t, cfg.Deployers)
	require.IsType(t, &infro.ArgoCDConfig{}, cfg.Deployers[0].Value)
	require.IsType(t, &infro.GithubConfig{}, cfg.VCS.Value)
}
