package injectionmodel

import (
	"encoding/json"
	"os"
	"testing"
)

func TestEmbeddedArtifactMatchesManifest(t *testing.T) {
	b, err := os.ReadFile("model.json")
	if err != nil {
		t.Fatal(err)
	}
	var meta struct {
		ModelID       string `json:"model_id"`
		FeatureSchema string `json:"feature_schema"`
		FeatureDim    int    `json:"feature_dim"`
		WeightsSHA256 string `json:"weights_sha256"`
	}
	if err := json.Unmarshal(b, &meta); err != nil {
		t.Fatal(err)
	}
	m, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if meta.ModelID != ID || meta.FeatureSchema != FeatureSchema || meta.FeatureDim != Dim {
		t.Fatalf("incompatible artifact metadata: %+v", meta)
	}
	if m.Dim != Dim || len(m.Weights) != Dim || meta.WeightsSHA256 != EmbeddedSHA256() {
		t.Fatal("embedded weights do not match the declared artifact")
	}
}
