package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultProviderConfigFile = "config/providers.yml"

// ProviderBootstrapEntry describes a provider that should be bootstrapped on startup.
type ProviderBootstrapEntry struct {
	Name                string
	Vendor              string
	BaseURL             string
	APIKey              string
	Active              bool
	Metadata            map[string]string
	AutoEnableNewModels bool
	SyncModels          bool
}

// ProviderBootstrapConfig maintains all configured provider sets.
type ProviderBootstrapConfig struct {
	sets map[string][]ProviderBootstrapEntry
}

// ProvidersForSet returns a copy of the providers defined for the requested set.
func (c *ProviderBootstrapConfig) ProvidersForSet(name string) []ProviderBootstrapEntry {
	if c == nil {
		return nil
	}
	set := strings.TrimSpace(name)
	if set == "" {
		set = "default"
	}
	list := c.sets[set]
	if len(list) == 0 {
		return nil
	}
	result := make([]ProviderBootstrapEntry, len(list))
	copy(result, list)
	return result
}

// LoadProviderBootstrapConfig parses the yaml file at the provided path.
func LoadProviderBootstrapConfig(path string) (*ProviderBootstrapConfig, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("provider config path is empty!!!")
	}

	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !filepath.IsAbs(cleanPath) {
			altPath := filepath.Clean(filepath.Join("services", "llm-api", cleanPath))
			altData, altErr := os.ReadFile(altPath)
			if altErr != nil {
				return nil, fmt.Errorf("read provider config %q: %w", altPath, altErr)
			}
			data = altData
			cleanPath = altPath
		} else {
			return nil, fmt.Errorf("read provider config %q: %w", cleanPath, err)
		}
	}

	var doc providerConfigDocument
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse provider config %q: %w", cleanPath, err)
	}

	if len(doc.Providers) == 0 {
		return nil, fmt.Errorf("provider config %q has no providers defined", cleanPath)
	}

	result := &ProviderBootstrapConfig{
		sets: make(map[string][]ProviderBootstrapEntry),
	}

	for rawSet, entries := range doc.Providers {
		setName := strings.TrimSpace(rawSet)
		if setName == "" || len(entries) == 0 {
			continue
		}
		for idx, entry := range entries {
			normalized, err := normalizeProviderEntry(entry)
			if err != nil {
				return nil, fmt.Errorf("providers.%s[%d]: %w", setName, idx, err)
			}
			result.sets[setName] = append(result.sets[setName], normalized)
		}
	}

	if len(result.sets) == 0 {
		return nil, fmt.Errorf("provider config %q has no valid provider entries", cleanPath)
	}

	return result, nil
}

type providerConfigDocument struct {
	Providers map[string][]providerConfigEntry `yaml:"providers"`
}

type providerConfigEntry struct {
	Name        string            `yaml:"name"`
	Type        string            `yaml:"type"`
	Vendor      string            `yaml:"vendor"`
	URL         string            `yaml:"url"`
	BaseURL     string            `yaml:"base_url"`
	APIKey      string            `yaml:"api_key"`
	Key         string            `yaml:"key"`
	Active      *bool             `yaml:"active"`
	Description string            `yaml:"description"`
	Metadata    map[string]string `yaml:"metadata"`
	AutoEnable  *bool             `yaml:"auto_enable_new_models"`
	SyncModels  *bool             `yaml:"sync_models"`
}

func normalizeProviderEntry(entry providerConfigEntry) (ProviderBootstrapEntry, error) {
	vendor := firstNonEmpty(entry.Type, entry.Vendor)
	vendor = strings.TrimSpace(vendor)
	if vendor == "" {
		return ProviderBootstrapEntry{}, errors.New("provider type is required")
	}

	baseURL := firstNonEmpty(entry.URL, entry.BaseURL)
	baseURL = strings.TrimSpace(os.ExpandEnv(baseURL))
	if baseURL == "" {
		return ProviderBootstrapEntry{}, errors.New("provider url is required")
	}

	name := strings.TrimSpace(entry.Name)
	if name == "" {
		name = fmt.Sprintf("%s Provider", strings.ToUpper(vendor))
	}
	name = os.ExpandEnv(name)

	apiKey := strings.TrimSpace(firstNonEmpty(entry.APIKey, entry.Key))
	if apiKey != "" {
		apiKey = os.ExpandEnv(apiKey)
	}

	active := true
	if entry.Active != nil {
		active = *entry.Active
	}

	autoEnable := true
	if entry.AutoEnable != nil {
		autoEnable = *entry.AutoEnable
	}

	syncModels := true
	if entry.SyncModels != nil {
		syncModels = *entry.SyncModels
	}

	metadata := cloneStringMap(entry.Metadata)
	if desc := strings.TrimSpace(os.ExpandEnv(entry.Description)); desc != "" {
		metadata = ensureStringMap(metadata)
		metadata["description"] = desc
	}
	metadata = ensureStringMap(metadata)
	metadata["auto_enable_new_models"] = strconv.FormatBool(autoEnable)
	if len(metadata) == 0 {
		metadata = nil
	}

	return ProviderBootstrapEntry{
		Name:                name,
		Vendor:              vendor,
		BaseURL:             baseURL,
		APIKey:              apiKey,
		Active:              active,
		Metadata:            metadata,
		AutoEnableNewModels: autoEnable,
		SyncModels:          syncModels,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		key := strings.TrimSpace(k)
		val := strings.TrimSpace(os.ExpandEnv(v))
		if key == "" || val == "" {
			continue
		}
		out[key] = val
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func ensureStringMap(in map[string]string) map[string]string {
	if in == nil {
		return make(map[string]string)
	}
	return in
}
