package injectionmodel

// ID identifies this classifier independently of application releases.
const ID = "saifety-prompt-injection"

// FeatureSchema versions the feature mapping required by the weight artifact.
// A change to feature indices or normalization requires a new schema.
const FeatureSchema = "hashed-char-ngram-v1"
