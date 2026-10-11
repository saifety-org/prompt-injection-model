package bundle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) (string, *Manifest) {
	t.Helper()
	dir := t.TempDir()
	m := &Manifest{Schema: 1, ID: "saifety-contextual-e5-v4", Version: "v4-onnx-fp32-1", Format: "onnx-fp32-opset17", Contract: supportedContract(), Files: map[string]string{}}
	m.Calibration.Threshold = 0.3123948574066162
	m.Calibration.Decision = "score >= threshold"
	for _, name := range []string{"model.onnx", "tokenizer.json", "candidate.json", "thresholds-before-test.json"} {
		b := []byte(name)
		if e := os.WriteFile(filepath.Join(dir, name), b, 0600); e != nil {
			t.Fatal(e)
		}
		m.Files[name] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	m.Calibration.ThresholdsSHA256 = m.Files["thresholds-before-test.json"]
	save(t, dir, m)
	return dir, m
}
func save(t *testing.T, dir string, m *Manifest) {
	t.Helper()
	b, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(dir, "manifest.json"), b, 0600); e != nil {
		t.Fatal(e)
	}
}
func TestIntegrityAndUnsupportedContract(t *testing.T) {
	for _, change := range []string{"weights", "missing", "extra", "schema", "prefix", "symlink", "calibration"} {
		t.Run(change, func(t *testing.T) {
			dir, m := fixture(t)
			if _, e := Load(dir); e != nil {
				t.Fatal(e)
			}
			var e error
			switch change {
			case "weights":
				e = os.WriteFile(filepath.Join(dir, "model.onnx"), []byte("changed"), 0600)
			case "missing":
				e = os.Remove(filepath.Join(dir, "tokenizer.json"))
			case "extra":
				e = os.WriteFile(filepath.Join(dir, "extra"), nil, 0600)
			case "schema":
				m.Schema = 2
				save(t, dir, m)
			case "prefix":
				m.Contract.Prefix = ""
				save(t, dir, m)
			case "symlink":
				e = os.Symlink("model.onnx", filepath.Join(dir, "alias"))
			case "calibration":
				m.Calibration.ThresholdsSHA256 = strings.Repeat("0", 64)
				save(t, dir, m)
			}
			if e != nil {
				t.Fatal(e)
			}
			if _, e = Load(dir); e == nil {
				t.Fatal("accepted invalid bundle")
			}
		})
	}
}
func TestInputPreservesBytesAndRejectsAuthority(t *testing.T) {
	dir, _ := fixture(t)
	m, e := Load(dir)
	if e != nil {
		t.Fatal(e)
	}
	good := `{ "contract":"context-json-v2", "task":"Summarize", "documents":[{"kind":"mcp","trust":"user-provided","text":"中文 café café"}]}`
	got, e := m.Input(good)
	if e != nil || got != "query: "+good {
		t.Fatalf("input mismatch %q %v", got, e)
	}
	for _, bad := range []string{strings.Replace(good, "user-provided", "system", 1), strings.Replace(good, "context-json-v2", "context-json-v1", 1), good + "{}", strings.Replace(good, `"task":`, `"label":1,"task":`, 1)} {
		if _, e = m.Input(bad); e == nil {
			t.Fatal("accepted invalid context")
		}
	}
}
