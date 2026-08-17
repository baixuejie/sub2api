package config

import (
	"errors"
	"path/filepath"
)

func WithBootstrapLock(paths Paths, operation func() error) error {
	if operation == nil {
		return errors.New("bootstrap operation is required")
	}
	if err := EnsurePrivateDir(paths.DataDir); err != nil {
		return err
	}
	return withFileLock(filepath.Join(paths.DataDir, "bootstrap"), operation)
}

func Apply(paths Paths, provider ProviderConfig, apiKey string) error {
	if err := validateProvider(provider); err != nil {
		return err
	}
	if apiKey == "" {
		return errors.New("API key must not be empty")
	}
	if err := EnsurePrivateDir(paths.DataDir); err != nil {
		return err
	}
	if err := EnsurePrivateDir(paths.DSHHome); err != nil {
		return err
	}
	return withFileLock(filepath.Join(paths.DataDir, "configuration"), func() error {
		if err := MergeSettings(paths.SettingsFile, provider); err != nil {
			return err
		}
		if err := MergeCredential(paths.CredentialsFile, provider.CredentialName, apiKey); err != nil {
			return err
		}
		return TightenPrivateTree(paths.DataDir, paths.DSHHome, paths.SettingsFile, paths.CredentialsFile)
	})
}
