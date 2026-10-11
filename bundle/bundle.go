// Package bundle validates research contextual ONNX bundles without loading a
// runtime. It does not change the default embedded native model.
package bundle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Contract is intentionally closed: new preprocessing requires a new schema.
type Contract struct {
	InputContract string            `json:"input_contract"`
	Serialization string            `json:"serialization"`
	Prefix        string            `json:"prefix"`
	Normalization string            `json:"normalization"`
	MaxTokens     int               `json:"max_tokens"`
	Overflow      string            `json:"overflow"`
	Labels        map[string]string `json:"labels"`
	AttackIndex   int               `json:"attack_index"`
	Pooling       string            `json:"pooling"`
	Inputs        []string          `json:"inputs"`
	Outputs       []string          `json:"outputs"`
}

// Manifest binds artifact bytes, input behavior and the frozen decision rule.
type Manifest struct {
	Schema          int      `json:"schema"`
	ID              string   `json:"id"`
	Version         string   `json:"version"`
	Format          string   `json:"format"`
	ProductionReady bool     `json:"production_ready"`
	Contract        Contract `json:"contract"`
	Calibration     struct {
		Threshold        float64 `json:"threshold"`
		Decision         string  `json:"decision"`
		Method           string  `json:"method"`
		ThresholdsSHA256 string  `json:"thresholds_sha256"`
	} `json:"calibration"`
	Provenance  json.RawMessage   `json:"provenance"`
	Limitations []string          `json:"limitations"`
	Files       map[string]string `json:"files"`
}

func supportedContract() Contract {
	return Contract{"context-json-v2", "preserve-validated-json-utf8-v1", "query: ", "tokenizer.json-only", 512, "error", map[string]string{"0": "benign", "1": "attack-on-agent"}, 1, "masked-mean-l2-normalized", []string{"input_ids", "attention_mask"}, []string{"logits"}}
}

// Load verifies the exact file set and hashes before any artifact is parsed by ML libraries.
// A manifest must itself come from a trusted pinned commit/hash; hashes are not signatures.
func Load(directory string) (*Manifest, error) {
	b, e := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if e != nil {
		return nil, e
	}
	var m Manifest
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if e = dec.Decode(&m); e != nil {
		return nil, e
	}
	if dec.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("trailing manifest data")
	}
	if m.Schema != 1 || m.ID != "saifety-contextual-e5-v4" || m.Version != "v4-onnx-fp32-1" || m.Format != "onnx-fp32-opset17" || m.ProductionReady || !reflect.DeepEqual(m.Contract, supportedContract()) {
		return nil, fmt.Errorf("unsupported contextual bundle schema/contract")
	}
	if math.IsNaN(m.Calibration.Threshold) || math.IsInf(m.Calibration.Threshold, 0) || m.Calibration.Threshold < 0 || m.Calibration.Threshold > 1 || m.Calibration.Decision != "score >= threshold" {
		return nil, fmt.Errorf("invalid calibration")
	}
	for _, name := range []string{"model.onnx", "tokenizer.json", "candidate.json", "thresholds-before-test.json"} {
		if _, ok := m.Files[name]; !ok {
			return nil, fmt.Errorf("missing artifact %s", name)
		}
	}
	if m.Files["thresholds-before-test.json"] != m.Calibration.ThresholdsSHA256 {
		return nil, fmt.Errorf("calibration binding mismatch")
	}
	seen := make(map[string]string)
	e = filepath.WalkDir(directory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink artifact")
		}
		if d.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(relative)
		if name == "manifest.json" {
			return nil
		}
		expected, ok := m.Files[name]
		if !ok {
			return fmt.Errorf("unexpected artifact %s", name)
		}
		decoded, err := hex.DecodeString(expected)
		if err != nil || len(decoded) != 32 {
			return fmt.Errorf("invalid SHA256")
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		h := sha256.New()
		_, err = io.Copy(h, f)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		actual := fmt.Sprintf("%x", h.Sum(nil))
		if actual != expected {
			return fmt.Errorf("artifact checksum mismatch: %s", name)
		}
		seen[name] = actual
		return nil
	})
	if e != nil {
		return nil, e
	}
	if !reflect.DeepEqual(seen, m.Files) {
		return nil, fmt.Errorf("bundle file set mismatch")
	}
	return &m, nil
}

// Input validates authority boundaries, preserves the supplied JSON bytes and
// adds the trained prefix. Changing JSON whitespace/order changes model input.
func (m *Manifest) Input(text string) (string, error) {
	var v struct {
		Contract  string `json:"contract"`
		Task      string `json:"task"`
		Documents []struct {
			Kind  string `json:"kind"`
			Trust string `json:"trust"`
			Text  string `json:"text"`
		} `json:"documents"`
	}
	dec := json.NewDecoder(strings.NewReader(text))
	dec.DisallowUnknownFields()
	if e := dec.Decode(&v); e != nil {
		return "", e
	}
	if dec.Decode(new(any)) != io.EOF {
		return "", fmt.Errorf("trailing context data")
	}
	if v.Contract != "context-json-v2" || strings.TrimSpace(v.Task) == "" || len(v.Documents) == 0 {
		return "", fmt.Errorf("invalid context contract")
	}
	for _, d := range v.Documents {
		if (d.Kind != "mcp" && d.Kind != "file" && d.Kind != "web") || (d.Trust != "untrusted" && d.Trust != "user-provided") || strings.TrimSpace(d.Text) == "" {
			return "", fmt.Errorf("invalid document or elevated trust")
		}
	}
	return m.Contract.Prefix + text, nil
}
