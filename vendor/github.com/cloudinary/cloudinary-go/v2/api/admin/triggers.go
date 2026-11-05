package admin

// Enables you to manage webhook notification triggers for your Cloudinary product environment.
//
// https://cloudinary.com/documentation/notifications

import (
	"context"

	"github.com/cloudinary/cloudinary-go/v2/api"
)

const (
	triggersEndpoint api.EndPoint = "triggers"
)

// Trigger represents a single webhook notification trigger.
type Trigger struct {
	ID                   string         `json:"id,omitempty"`
	URI                  string         `json:"uri,omitempty"`
	EventType            string         `json:"event_type,omitempty"`
	Additive             bool           `json:"additive"`
	Filter               map[string]any `json:"filter,omitempty"`
	FilterLanguage       string         `json:"filter_language,omitempty"`
	PayloadTemplate      map[string]any `json:"payload_template,omitempty"`
	AuthScheme           string         `json:"auth_scheme,omitempty"`
	ProductEnvironmentID string         `json:"product_environment_id,omitempty"`
	URIType              string         `json:"uri_type,omitempty"`
	CreatedAt            string         `json:"created_at,omitempty"`
	UpdatedAt            string         `json:"updated_at,omitempty"`
}

// ListTriggersParams are the parameters for ListTriggers.
type ListTriggersParams struct{}

// ListTriggersResult is the result of ListTriggers.
type ListTriggersResult struct {
	Triggers []Trigger     `json:"triggers"`
	Error    api.ErrorResp `json:"error,omitempty"`
}

// ListTriggers lists all webhook notification triggers for the product environment.
func (a *API) ListTriggers(ctx context.Context, params ListTriggersParams) (*ListTriggersResult, error) {
	res := &ListTriggersResult{}
	_, err := a.get(ctx, triggersEndpoint, params, res)
	return res, err
}

// CreateTriggerParams are the parameters for CreateTrigger.
type CreateTriggerParams struct {
	URI             string         `json:"uri"`
	EventType       string         `json:"event_type"`
	Additive        bool           `json:"additive,omitempty"`
	Filter          map[string]any `json:"filter,omitempty"`
	FilterLanguage  string         `json:"filter_language,omitempty"`
	PayloadTemplate map[string]any `json:"payload_template,omitempty"`
	AuthScheme      string         `json:"auth_scheme,omitempty"`
}

// CreateTriggerResult is the result of CreateTrigger.
type CreateTriggerResult struct {
	Trigger
	Error api.ErrorResp `json:"error,omitempty"`
}

// CreateTrigger creates a new webhook notification trigger.
func (a *API) CreateTrigger(ctx context.Context, params CreateTriggerParams) (*CreateTriggerResult, error) {
	res := &CreateTriggerResult{}
	_, err := a.post(ctx, triggersEndpoint, params, res)
	return res, err
}

// UpdateTriggerParams are the parameters for UpdateTrigger.
type UpdateTriggerParams struct {
	TriggerID       string         `json:"-"`
	URI             string         `json:"uri,omitempty"`
	EventType       string         `json:"event_type,omitempty"`
	Additive        *bool          `json:"additive,omitempty"`
	Filter          map[string]any `json:"filter,omitempty"`
	FilterLanguage  string         `json:"filter_language,omitempty"`
	PayloadTemplate map[string]any `json:"payload_template,omitempty"`
	AuthScheme      string         `json:"auth_scheme,omitempty"`
}

// UpdateTriggerResult is the result of UpdateTrigger.
type UpdateTriggerResult struct {
	Trigger
	Error api.ErrorResp `json:"error,omitempty"`
}

// UpdateTrigger updates an existing webhook notification trigger.
func (a *API) UpdateTrigger(ctx context.Context, params UpdateTriggerParams) (*UpdateTriggerResult, error) {
	res := &UpdateTriggerResult{}
	_, err := a.put(ctx, api.BuildPath(triggersEndpoint, params.TriggerID), params, res)
	return res, err
}

// DeleteTriggerParams are the parameters for DeleteTrigger.
type DeleteTriggerParams struct {
	TriggerID string `json:"-"`
}

// DeleteTriggerResult is the result of DeleteTrigger.
type DeleteTriggerResult struct {
	Message string        `json:"message,omitempty"`
	Error   api.ErrorResp `json:"error,omitempty"`
}

// DeleteTrigger deletes a webhook notification trigger by its ID.
func (a *API) DeleteTrigger(ctx context.Context, params DeleteTriggerParams) (*DeleteTriggerResult, error) {
	res := &DeleteTriggerResult{}
	_, err := a.delete(ctx, api.BuildPath(triggersEndpoint, params.TriggerID), params, res)
	return res, err
}
