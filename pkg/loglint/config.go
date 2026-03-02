package loglint

import (
	"encoding/json"
	"fmt"
	"os"
)

var defaultSensitiveKeywords = []string{
	"password",
	"passwd",
	"pwd",
	"apikey",
	"secret",
	"token",
}

type Config struct {
	CheckLowercaseStart *bool    `json:"check_lowercase_start,omitempty"`
	CheckEnglishOnly    *bool    `json:"check_english_only,omitempty"`
	CheckSpecialChars   *bool    `json:"check_special_chars,omitempty"`
	CheckSensitiveData  *bool    `json:"check_sensitive_data,omitempty"`
	SensitiveKeywords   []string `json:"sensitive_keywords,omitempty"`
}

type effectiveConfig struct {
	CheckLowercaseStart bool
	CheckEnglishOnly    bool
	CheckSpecialChars   bool
	CheckSensitiveData  bool
	SensitiveKeywords   []string
}

func defaultConfig() effectiveConfig {
	return effectiveConfig{
		CheckLowercaseStart: true,
		CheckEnglishOnly:    true,
		CheckSpecialChars:   true,
		CheckSensitiveData:  true,
		SensitiveKeywords:   append([]string(nil), defaultSensitiveKeywords...),
	}
}

func loadConfigFile(path string) (*Config, error) {
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("decode config %q: %w", path, err)
	}

	return &cfg, nil
}

func mergeConfig(base effectiveConfig, overrides ...*Config) effectiveConfig {
	cfg := effectiveConfig{
		CheckLowercaseStart: base.CheckLowercaseStart,
		CheckEnglishOnly:    base.CheckEnglishOnly,
		CheckSpecialChars:   base.CheckSpecialChars,
		CheckSensitiveData:  base.CheckSensitiveData,
		SensitiveKeywords:   append([]string(nil), base.SensitiveKeywords...),
	}

	for _, override := range overrides {
		if override == nil {
			continue
		}

		if override.CheckLowercaseStart != nil {
			cfg.CheckLowercaseStart = *override.CheckLowercaseStart
		}
		if override.CheckEnglishOnly != nil {
			cfg.CheckEnglishOnly = *override.CheckEnglishOnly
		}
		if override.CheckSpecialChars != nil {
			cfg.CheckSpecialChars = *override.CheckSpecialChars
		}
		if override.CheckSensitiveData != nil {
			cfg.CheckSensitiveData = *override.CheckSensitiveData
		}
		if override.SensitiveKeywords != nil {
			cfg.SensitiveKeywords = append([]string(nil), override.SensitiveKeywords...)
		}
	}

	return cfg
}

func ConfigFromAny(raw any) (*Config, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case Config:
		cfg := v
		return &cfg, nil
	case *Config:
		return v, nil
	case map[string]any:
		return configFromMap(v)
	default:
		return nil, fmt.Errorf("unsupported config type %T", raw)
	}
}

func configFromMap(raw map[string]any) (*Config, error) {
	var cfg Config

	if rawConfigPath, ok := raw["config"]; ok {
		path, ok := rawConfigPath.(string)
		if !ok {
			return nil, fmt.Errorf("config must be a string")
		}
		fileConfig, err := loadConfigFile(path)
		if err != nil {
			return nil, err
		}
		applyConfigOverride(&cfg, fileConfig)
	}

	for key, value := range raw {
		switch key {
		case "check_lowercase_start":
			b, ok := value.(bool)
			if !ok {
				return nil, fmt.Errorf("%s must be a boolean", key)
			}
			cfg.CheckLowercaseStart = &b
		case "check_english_only":
			b, ok := value.(bool)
			if !ok {
				return nil, fmt.Errorf("%s must be a boolean", key)
			}
			cfg.CheckEnglishOnly = &b
		case "check_special_chars":
			b, ok := value.(bool)
			if !ok {
				return nil, fmt.Errorf("%s must be a boolean", key)
			}
			cfg.CheckSpecialChars = &b
		case "check_sensitive_data":
			b, ok := value.(bool)
			if !ok {
				return nil, fmt.Errorf("%s must be a boolean", key)
			}
			cfg.CheckSensitiveData = &b
		case "sensitive_keywords":
			values, err := stringSliceFromAny(value)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
			cfg.SensitiveKeywords = values
		case "config":
			continue
		}
	}

	return &cfg, nil
}

func stringSliceFromAny(value any) ([]string, error) {
	switch v := value.(type) {
	case []string:
		return append([]string(nil), v...), nil
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("must be an array of strings")
			}
			out = append(out, s)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("must be an array of strings")
	}
}

func applyConfigOverride(dst *Config, src *Config) {
	if dst == nil || src == nil {
		return
	}

	if src.CheckLowercaseStart != nil {
		value := *src.CheckLowercaseStart
		dst.CheckLowercaseStart = &value
	}
	if src.CheckEnglishOnly != nil {
		value := *src.CheckEnglishOnly
		dst.CheckEnglishOnly = &value
	}
	if src.CheckSpecialChars != nil {
		value := *src.CheckSpecialChars
		dst.CheckSpecialChars = &value
	}
	if src.CheckSensitiveData != nil {
		value := *src.CheckSensitiveData
		dst.CheckSensitiveData = &value
	}
	if src.SensitiveKeywords != nil {
		dst.SensitiveKeywords = append([]string(nil), src.SensitiveKeywords...)
	}
}
