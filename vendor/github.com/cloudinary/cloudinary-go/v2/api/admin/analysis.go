package admin

import (
	"context"
	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	analysis api.EndPoint = "analysis"
	analyze  api.EndPoint = "analyze"
	uri      api.EndPoint = "uri"
)

// AnalyzeParams are the parameters for Analyze.
type AnalyzeParams struct {
	// The URI of the asset to analyze
	Uri          string                       `json:"uri,omitempty"`
	AnalysisType string                       `json:"analysis_type,omitempty"`
	Parameters   *AnalyzeUriRequestParameters `json:"parameters,omitempty"`
}

// AnalyzeUriRequestParameters struct for AnalyzeUriRequestParameters
type AnalyzeUriRequestParameters struct {
	Custom *CustomParameters `json:"custom,omitempty"`
}

// CustomParameters struct for CustomParameters
type CustomParameters struct {
	ModelName    string `json:"model_name,omitempty"`
	ModelVersion int    `json:"model_version,omitempty"`
}

/*
Analyze Analyzes an asset with the requested analysis type.

Currently supports the following analysis options:
* [Google tagging](https://cloudinary.com/documentation/google_auto_tagging_addon)
* [Captioning](https://cloudinary.com/documentation/cloudinary_ai_content_analysis_addon#ai_based_image_captioning)
* [Cld Fashion](https://cloudinary.com/documentation/cloudinary_ai_content_analysis_addon#supported_content_aware_detection_models)
* [Coco](https://cloudinary.com/documentation/cloudinary_ai_content_analysis_addon#supported_content_aware_detection_models)
* [Lvis](https://cloudinary.com/documentation/cloudinary_ai_content_analysis_addon#supported_content_aware_detection_models)
* [Unidet](https://cloudinary.com/documentation/cloudinary_ai_content_analysis_addon#supported_content_aware_detection_models)
* [Human Anatomy](https://cloudinary.com/documentation/cloudinary_ai_content_analysis_addon#supported_content_aware_detection_models)
* [Cld Text](https://cloudinary.com/documentation/cloudinary_ai_content_analysis_addon#supported_content_aware_detection_models)
* [Shop Classifier](https://cloudinary.com/documentation/cloudinary_ai_content_analysis_addon#supported_content_aware_detection_models)
* Custom
*/
func (a *API) Analyze(ctx context.Context, params AnalyzeParams) (*AnalyzeResult, error) {
	v2APICtx := context.WithValue(ctx, "api_version", "2")
	res := &AnalyzeResult{}
	_, err := a.post(v2APICtx, api.BuildPath(analysis, analyze, uri), params, res)

	return res, err
}

// AnalyzeResult is the result of Analyze.
type AnalyzeResult struct {
	Data      AnalysisPayload `json:"data"`
	RequestId string          `json:"request_id"`
	Error     api.ErrorResp   `json:"error,omitempty"`
	Response  interface{}
}

// AnalysisPayload struct for AnalysisPayload
type AnalysisPayload struct {
	Entity   string                 `json:"entity"`
	Analysis map[string]interface{} `json:"analysis"`
}
