package scenario

import (
	"encoding/json"
	"os"

	"github.com/solguardlabs/northstardtl/src/api"
)

func LoadFile(path string) (Definition, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Definition{}, err
	}
	var definition Definition
	if err := json.Unmarshal(raw, &definition); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func LoadBootstrapFile(path string) (api.Bootstrap, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return api.Bootstrap{}, err
	}
	var bootstrap api.Bootstrap
	if err := json.Unmarshal(raw, &bootstrap); err != nil {
		return api.Bootstrap{}, err
	}
	return bootstrap, nil
}
