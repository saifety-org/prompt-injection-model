# sAIfety prompt-injection model

The native Go prompt-injection classifier used by
[sAIfety](https://github.com/saifety-org/sAIfety). This repository versions
ready weights together with their feature extraction and inference code.
It is separate from the application's scanner, MCP proxy and policies, and
from a possible future personal-data/PII model.

The classifier uses logistic regression over hashed character n-grams.
It scores text locally and embeds its weights into the consuming Go binary;
it does not download models or call an external inference service.

## Use

```go
import injectionmodel "github.com/saifety-org/prompt-injection-model"

model, err := injectionmodel.Default()
if err != nil {
    // Handle the invalid embedded artifact.
    return
}
score := model.Score("Text from an untrusted source")
```

Higher scores indicate an attack on the agent. Scores are not a guarantee of
calibrated probabilities. The application chooses thresholds and actions;
this module only supplies scores. `Load`, `Features`, `Model` and `Dim` also
support candidate training and inference in the laboratory.

## Repositories and releases

- `sAIfety`: application, scanner, rules, proxy, hooks and policy.
- `prompt-injection-model`: production weights, feature schema, inference
  and artifact compatibility tests.
- [lab](https://github.com/saifety-org/lab): datasets, training, examples and
  evaluations. Training a candidate does not replace this artifact automatically.

Consumers pin an exact module version in `go.mod`. The model's `go:embed`
includes weights at build time, producing a single application binary that
can run offline. A newer model version requires an explicit dependency-update
PR in the application; deployment does not fetch `latest`.

Model versions are independent of application versions. Release provenance
must record both versions, `FeatureSchema` and `EmbeddedSHA256()`. This module
has no dependency on the application or laboratory, avoiding dependency cycles.

## Artifact and limitations

[model.json](model.json) records the original source commit, file hashes,
feature schema and weight hash. The current weights are copied byte-for-byte
from sAIfety; extraction does not retrain or change predictions.

The model was trained on synthetic attack/benign examples, public corpora and
local benign command examples. See the laboratory's
[attribution](https://github.com/saifety-org/lab/blob/main/datasets/training/ATTRIBUTION.md).
Historical metrics inside `weights.json` use a same-source held-out split and
are not independent quality estimates. The
[external comparison](https://github.com/saifety-org/lab/blob/main/docs/model-comparison-results.md)
exposes substantial misses and false positives. Some benchmark labels require
source/task context absent from isolated text scoring. Evaluate classifier
scores and complete agent protection separately; this artifact is experimental.

It does not recognize or redact PII, parse commands, decode hidden payloads,
apply trust policy, or decide whether an agent should execute a tool. Those
are application responsibilities or separate models.

## Development

Go 1.26+ is required. There are no external Go dependencies.

```sh
make test vet build
```

The manifest test checks the embedded artifact against the declared checksum,
dimensions and feature schema. Example-based tests are focused regression checks,
not a benchmark or evidence of general protection quality.

## CI checks

Pull requests and pushes to `main` run three required checks: `lint`, `test`,
`build`. Reproduce them from this repository with Go from `go.mod` and a C
compiler for the ONNX build where applicable:

```sh
make lint-install          # golangci-lint v2.14.0; installs only into ./bin
make lint                 # gofmt (read-only), go vet, configured Go linters
make ci-test              # unit/regression tests with race detector
make ci-build             # all supported build variants
```

`GOWORK=off` and `-mod=readonly` prevent local workspace overrides or implicit
module edits. Actions are pinned by commit; the linter version is pinned in
both CI and Makefile. Checks have timeouts and newer runs cancel stale runs
on the same PR. CI does not download inference models, train candidates or
run full benchmarks.

The linter uses the standard checks (`errcheck`, `govet`, `ineffassign`,
`staticcheck`, `unused`) without automatic fixes. No linter exclusions are configured.

Golden feature-vector tests detect incompatible extractor changes under the same
feature-schema name; the manifest test validates embedded weights and dimensions.
