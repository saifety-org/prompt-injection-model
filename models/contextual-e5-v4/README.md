# Contextual E5 v4 research artifact

`manifest.json` versions the local `v4-onnx-fp32-1` bundle by exact SHA256.
It is **not the default model and not approved for production**. The existing
embedded native model and `model.json` are unchanged.

The accepted lab-py v4 checkpoint fine-tunes multilingual-e5-small's BERT
encoder and binary head end to end. It uses the checkpoint's XLM-R Unigram
**tokenizer**, masked mean pooling and L2 normalization. ONNX contains the
encoder, pooling, normalization and head; output is two FP32 logits.

Input is validated `context-json-v2`: task plus ordered documents with kind,
trust and text. Only `untrusted` and `user-provided` are accepted; neither
confers system authority. Preserve the supplied JSON bytes and prepend
`query: `. Whitespace, field order and escaping affect tokens: do not silently
reserialize. No case folding, translation or additional Unicode normalization.
Apply tokenizer.json, special tokens included, at most 512 tokens; overflow
is an error. Score is softmax(logits)[1]; alert when score >=
0.3123948574066162, fixed on v4 validation before test inference.

## Obtain and verify

Export from the accepted local safetensors checkpoint with the locked lab-py
workflow in [contextual-onnx.md](https://github.com/saifety-org/lab-py/blob/main/docs/contextual-onnx.md).
The exporter refuses overwrite and copies the tokenizer bytes unchanged.
Compare the generated manifest with this pinned manifest, then run:

```sh
go run ./cmd/verify-context-bundle /path/to/bundle
```

The verifier checks the closed input contract, version, calibration binding,
exact artifact set and every checksum. Unsupported schemas, missing/changed
files and symlinks fail before loading ML libraries. Pin this manifest's
trusted repository commit/hash: self-declared checksums are not signatures.

The 487 MB bundle is stored locally under ignored `artifacts/`, not in Git.
There is **no hosted binary download, release tag or automatic fetch** in this
change. This repository holds the artifact identity and Go verification;
training/export remains lab-py and Go research inference remains lab.
The manifest deliberately marks the derived-weight license unresolved.
Base checkpoint: MIT, immutable upstream revision in provenance. Training
corpus: NOASSERTION, synthetic model-authored/reviewed data. These are distinct
license/provenance facts; do not infer a distribution license for derived weights.

## Validation and limits

Actual PyTorch/Python ONNX/Go ONNX parity on 560 already observed v4 test inputs
plus 15 fixed multilingual/Unicode/510–512 token probes: exact token IDs and
all decisions agree; max FP32 score differences 7.74e-6 (export), 1.52e-6
(Python/Go with different pinned runtime binaries), tolerance 1e-5 fixed before
checking. [Full report](https://github.com/saifety-org/lab-py/blob/main/reports/contextual-onnx-v4.json).
This is numerical export validation, not a new blind quality benchmark.

Observed synthetic test FPR remains 1.43%, above the 1% goal. No retraining,
threshold retuning, app adapter or production promotion is included.
The lab tokenizer returns IDs only, with no original-text offsets. Its narrow
XLM-R compatibility path fixes known upstream Unicode normalization and long
piece defects; tested parity does not prove equivalence on all Unicode text.
