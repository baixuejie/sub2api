package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

var credentialNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

var managedProviderProtocols = map[string]map[string]struct{}{
	"sub2api-openai":      {"openai-responses": {}},
	"sub2api-anthropic":   {"anthropic-messages": {}},
	"sub2api-grok":        {"openai-responses": {}},
	"sub2api-gemini":      {"openai-completions": {}},
	"sub2api-antigravity": {"anthropic-messages": {}, "openai-completions": {}},
	"sub2api-composite":   {"openai-completions": {}},
}

type ProviderConfig struct {
	Route          string
	DisplayName    string
	Protocol       string
	BaseURL        string
	CredentialName string
	ModelID        string
	ModelName      string
	ContextWindow  int
	MaxTokens      int
}

func MergeSettings(filename string, provider ProviderConfig) error {
	if err := validateProvider(provider); err != nil {
		return err
	}
	return withFileLock(filename, func() error {
		root, err := readDocument(filename)
		if err != nil {
			return err
		}
		mapping := documentMapping(root)
		llm, err := ensureMappingValue(mapping, "llm-pi-ai")
		if err != nil {
			return err
		}
		providers, err := ensureMappingValue(llm, "providers")
		if err != nil {
			return err
		}
		if hasCredentialConflict(providers, provider.CredentialName) {
			return errors.New("SUB2API_API_KEY is already used by a non-Sub2API provider")
		}
		removeManagedProviders(providers)
		setMappingValue(providers, provider.Route, providerNode(provider))
		model := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		setScalar(model, "provider", provider.Route)
		setScalar(model, "model", provider.ModelID)
		setMappingValue(mapping, "agent-default-model", model)
		text, err := encodeDocument(root)
		if err != nil {
			return err
		}
		if err := backupExisting(filename); err != nil {
			return err
		}
		return writeAtomic(filename, text)
	})
}

func MergeCredential(filename, name, value string) error {
	if name != "SUB2API_API_KEY" || !credentialNamePattern.MatchString(name) {
		return errors.New("credential_name must be SUB2API_API_KEY")
	}
	if value == "" {
		return errors.New("API key must not be empty")
	}
	return withFileLock(filename, func() error {
		root, err := readDocument(filename)
		if err != nil {
			return err
		}
		mapping := documentMapping(root)
		setScalar(mapping, name, value)
		text, err := encodeDocument(root)
		if err != nil {
			return err
		}
		if err := backupExisting(filename); err != nil {
			return err
		}
		return writeAtomic(filename, text)
	})
}

func readDocument(filename string) (*yaml.Node, error) {
	data, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	root := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
	if os.IsNotExist(err) || len(bytes.TrimSpace(data)) == 0 {
		return root, nil
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(false)
	if err := decoder.Decode(root); err != nil {
		return nil, fmt.Errorf("parse %s: %w", filename, err)
	}
	if len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("YAML document root must be a mapping")
	}
	return root, nil
}

func documentMapping(root *yaml.Node) *yaml.Node {
	if len(root.Content) == 0 {
		root.Content = append(root.Content, &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"})
	}
	return root.Content[0]
}

func ensureMappingValue(mapping *yaml.Node, key string) (*yaml.Node, error) {
	if existing := mappingValue(mapping, key); existing != nil {
		if existing.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("existing %s namespace must be a mapping", key)
		}
		return existing, nil
	}
	value := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	setMappingValue(mapping, key, value)
	return value, nil
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func isManagedProviderRoute(route string) bool {
	if route == "sub2api" {
		return true
	}
	_, exists := managedProviderProtocols[route]
	return exists
}

func hasCredentialConflict(providers *yaml.Node, credentialName string) bool {
	for i := 0; i+1 < len(providers.Content); i += 2 {
		key := providers.Content[i].Value
		if isManagedProviderRoute(key) {
			continue
		}
		provider := providers.Content[i+1]
		if provider.Kind != yaml.MappingNode {
			continue
		}
		credential := mappingValue(provider, "apiKeyEnv")
		if credential != nil && credential.Kind == yaml.ScalarNode && credential.Value == credentialName {
			return true
		}
	}
	return false
}

func removeManagedProviders(providers *yaml.Node) {
	filtered := make([]*yaml.Node, 0, len(providers.Content))
	for i := 0; i+1 < len(providers.Content); i += 2 {
		key := providers.Content[i].Value
		if isManagedProviderRoute(key) {
			continue
		}
		filtered = append(filtered, providers.Content[i], providers.Content[i+1])
	}
	providers.Content = filtered
}

func setMappingValue(mapping *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1] = value
			return
		}
	}
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
}

func setScalar(mapping *yaml.Node, key string, value any) {
	node := &yaml.Node{}
	_ = node.Encode(value)
	setMappingValue(mapping, key, node)
}

func providerNode(provider ProviderConfig) *yaml.Node {
	result := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	setScalar(result, "displayName", provider.DisplayName)
	setScalar(result, "apiKeyEnv", provider.CredentialName)
	setScalar(result, "api", provider.Protocol)
	setScalar(result, "baseURL", provider.BaseURL)
	model := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	setScalar(model, "id", provider.ModelID)
	setScalar(model, "name", provider.ModelName)
	setScalar(model, "contextWindow", provider.ContextWindow)
	setScalar(model, "maxTokens", provider.MaxTokens)
	models := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{model}}
	setMappingValue(result, "models", models)
	return result
}

func encodeDocument(root *yaml.Node) (string, error) {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(root); err != nil {
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func validateProvider(provider ProviderConfig) error {
	if provider.Route == "" || provider.DisplayName == "" || provider.Protocol == "" || provider.BaseURL == "" || provider.ModelID == "" || provider.ModelName == "" || provider.ContextWindow <= 0 || provider.MaxTokens <= 0 {
		return errors.New("provider configuration is incomplete")
	}
	if provider.CredentialName != "SUB2API_API_KEY" || !credentialNamePattern.MatchString(provider.CredentialName) {
		return errors.New("credential_name must be SUB2API_API_KEY")
	}
	protocols, exists := managedProviderProtocols[provider.Route]
	if !exists {
		return errors.New("provider route is not managed by Sub2API")
	}
	if _, exists := protocols[provider.Protocol]; !exists {
		return errors.New("provider protocol is not allowed for its route")
	}
	return nil
}
