package config

import (
	"context"
	"errors"
	"os"

	infisical "github.com/infisical/go-sdk"
)

type InfisicalConfig struct {
	Enabled                bool
	SiteURL                string
	ClientID               string
	ClientSecret           string
	ProjectID              string
	ProjectSlug            string
	Environment            string
	SecretPath             string
	IncludeImports         bool
	Recursive              bool
	ExpandSecretReferences bool
	Override               bool
	OrganizationSlug       string
}

type infisicalSecretLoader func(ctx context.Context, cfg InfisicalConfig) (map[string]string, error)

var loadInfisicalSecrets infisicalSecretLoader = fetchInfisicalSecrets

func fetchInfisicalSecrets(ctx context.Context, cfg InfisicalConfig) (map[string]string, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	client := infisical.NewInfisicalClient(ctx, infisical.Config{
		SiteUrl:    cfg.SiteURL,
		SilentMode: true,
	})

	auth := client.Auth()
	if cfg.OrganizationSlug != "" {
		auth = auth.WithOrganizationSlug(cfg.OrganizationSlug)
	}
	if _, err := auth.UniversalAuthLogin(cfg.ClientID, cfg.ClientSecret); err != nil {
		return nil, err
	}

	result, err := client.Secrets().ListSecrets(infisical.ListSecretsOptions{
		ProjectID:              cfg.ProjectID,
		ProjectSlug:            cfg.ProjectSlug,
		Environment:            cfg.Environment,
		SecretPath:             cfg.SecretPath,
		IncludeImports:         cfg.IncludeImports,
		Recursive:              cfg.Recursive,
		ExpandSecretReferences: cfg.ExpandSecretReferences,
	})
	if err != nil {
		return nil, err
	}

	secrets := make(map[string]string, len(result.Secrets))
	for _, secret := range result.Secrets {
		secrets[secret.SecretKey] = secret.SecretValue
	}

	return secrets, nil
}

func (cfg InfisicalConfig) validate() error {
	if cfg.ClientID == "" {
		return errors.New("INFISICAL_CLIENT_ID is required")
	}
	if cfg.ClientSecret == "" {
		return errors.New("INFISICAL_CLIENT_SECRET is required")
	}
	if cfg.ProjectID == "" && cfg.ProjectSlug == "" {
		return errors.New("INFISICAL_PROJECT_ID or INFISICAL_PROJECT_SLUG is required")
	}
	if cfg.Environment == "" {
		return errors.New("INFISICAL_ENVIRONMENT is required")
	}
	return nil
}

func applyInfisicalSecrets(secrets map[string]string, override bool) {
	for key, value := range secrets {
		if !override {
			currentValue, ok := os.LookupEnv(key)
			if ok && currentValue != "" {
				continue
			}
		}
		os.Setenv(key, value)
	}
}
