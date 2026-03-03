// Package api provides data models for UniFi Site Manager API
package api

import "encoding/json"

// APIResponse is the base response structure
type APIResponse struct {
	Code       string          `json:"code"`
	Data       json.RawMessage `json:"data"`
	HTTPStatus int             `json:"httpStatusCode"`
	TraceID    string          `json:"traceId"`
	NextToken  string          `json:"nextToken,omitempty"`
}

// Site represents a UniFi site
type Site struct {
	ID          string          `json:"_id"`
	Name        string          `json:"name"`
	Description string          `json:"desc"`
	HostID      string          `json:"hostId"`
	Meta        SiteMeta        `json:"meta"`
	Statistics  json.RawMessage `json:"statistics,omitempty"`
}

// SiteMeta contains site metadata
type SiteMeta struct {
	Name       string `json:"name"`
	GatewayMAC string `json:"gateway_mac"`
	NetworkID  string `json:"network_id"`
}

// SitesResponse wraps the list of sites
type SitesResponse struct {
	Code       string `json:"code"`
	Data       []Site `json:"data"`
	HTTPStatus int    `json:"httpStatusCode"`
	TraceID    string `json:"traceId"`
	NextToken  string `json:"nextToken,omitempty"`
}

// SiteResponse wraps a single site
type SiteResponse struct {
	Code       string `json:"code"`
	Data       Site   `json:"data"`
	HTTPStatus int    `json:"httpStatusCode"`
	TraceID    string `json:"traceId"`
}

// WhoamiResponse contains authenticated user information
type WhoamiResponse struct {
	Code       string   `json:"code"`
	Data       UserInfo `json:"data"`
	HTTPStatus int      `json:"httpStatusCode"`
	TraceID    string   `json:"traceId"`
}

// UserInfo represents the authenticated user
type UserInfo struct {
	ID        string `json:"_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	IsOwner   bool   `json:"isOwner"`
}
