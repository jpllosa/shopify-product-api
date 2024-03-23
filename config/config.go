package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	Port    int     `json:"port"`
	Shopify Shopify `json:"shopify"`
}

type Shopify struct {
	Endpoint          string `json:"endpoint"`
	AccessToken       string `json:"accessToken"`
	TargetEndpoint    string `json:"targetEndpoint"`
	TargetAccessToken string `json:"targetAccessToken"`
}

func Load(filePath string) (Config, error) {
	config, err := unmarshalFile[Config](filePath)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func unmarshalFile[T any](filePath string) (T, error) {
	pointer := new(T)

	f, err := os.Open(filePath)
	if err != nil {
		return *pointer, fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	jsonBytes, err := io.ReadAll(f)
	if err != nil {
		return *pointer, fmt.Errorf("reading config file: %w", err)
	}

	err = json.Unmarshal(jsonBytes, &pointer)
	if err != nil {
		return *pointer, fmt.Errorf("unmarshalling config file: %w", err)
	}

	return *pointer, nil
}
