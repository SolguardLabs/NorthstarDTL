package report

import (
	"encoding/json"
	"io"

	"github.com/solguardlabs/northstardtl/src/scenario"
)

func WriteScenarioResult(w io.Writer, result scenario.Result) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
