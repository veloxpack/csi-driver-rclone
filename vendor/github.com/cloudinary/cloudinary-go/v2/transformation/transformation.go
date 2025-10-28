// Package transformation defines Cloudinary Transformation.
package transformation

// This is a placeholder for a Transformation struct.

// Action represents a single transformation action. Consist of qualifiers.
type Action = map[string]interface{}

// Transformation is the asset transformation. Consists of Actions.
type Transformation = []Action

// RawTransformation is the raw (free form) transformation string.
type RawTransformation = string
