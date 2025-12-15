package models

import "time"

// Server represents a VPN server/node
type Server struct {
	CountryCode  string `json:"countryCode" db:"country_code"`
	CityName     string `json:"cityName" db:"city_name"`
	ExtName      string `json:"extName,omitempty" db:"ext_name"`
	Endpoint     string `json:"endpoint" db:"endpoint"`
	InboundType  string `json:"inboundType" db:"inbound_type"`
	Active       bool   `json:"active" db:"active"`
	CreatedAt    time.Time `json:"createdAt,omitempty" db:"created_at"`
}

// UserNode tracks which user was created on which node
type UserNode struct {
	UserID    string    `db:"user_id"`
	Endpoint  string    `db:"endpoint"`
	Inbound   string    `db:"inbound"`
	CreatedAt time.Time `db:"created_at"`
}

// CreateUserResponse is the response for /api/create
type CreateUserResponse struct {
	UUID    string         `json:"uuid"`
	Configs []ConfigItem   `json:"configs"`
}

// ConfigItem represents a single config for a country
type ConfigItem struct {
	CountryCode string `json:"countryCode"`
	Config      string `json:"config"`
}

// XrayUserResponse is the response from xray-service when creating a user
type XrayUserResponse struct {
	ID             string                 `json:"id"`
	Inbound        string                 `json:"inbound"`
	ConnectionConfig map[string]interface{} `json:"connection_config"`
}

// XrayInboundInfo represents inbound configuration from xray-service
type XrayInboundInfo struct {
	Inbound        string                 `json:"inbound"`
	ConnectionConfig map[string]interface{} `json:"connection_config"`
}

// XrayInboundsResponse is the response from /inbound endpoint
type XrayInboundsResponse struct {
	Inbounds []XrayInboundInfo `json:"inbounds"`
}

// XrayUser represents a user from xray-service
type XrayUser struct {
	ID      string `json:"id"`
	Inbound string `json:"inbound"`
}

// XrayUsersResponse is the response from /user endpoint
type XrayUsersResponse struct {
	Users []XrayUser `json:"users"`
}

// ServerRequest for POST/PUT/DELETE /api/servers
type ServerRequest struct {
	Servers []Server `json:"servers"`
}

// DeleteUserRequest for DELETE /api/deleteUser
type DeleteUserRequest struct {
	ID       string `json:"id"`
	Endpoint string `json:"endpoint,omitempty"` // Optional: for backward compatibility
}

// NodeRequest for POST /api/getUsers and /api/getInbounds
type NodeRequest struct {
	Endpoint string `json:"endpoint"`
}

// StatsResponse for GET /api/stats
type StatsResponse struct {
	TotalUsers  int                `json:"totalUsers"`
	Nodes       []NodeStatsItem    `json:"nodes"`
	ByProtocol  map[string]int     `json:"byProtocol"`
}

// NodeStatsItem represents statistics for a single node
type NodeStatsItem struct {
	Endpoint    string `json:"endpoint"`
	CountryCode string `json:"countryCode"`
	CityName    string `json:"cityName"`
	ExtName     string `json:"extName,omitempty"`
	InboundType string `json:"inboundType"`
	Active      bool   `json:"active"`
	UsersCount  int    `json:"usersCount"`
}

// UsersCountResponse for GET /api/users/count
type UsersCountResponse struct {
	Count int `json:"count"`
}

// UserListResponse for GET /api/users
type UserListResponse struct {
	Users []UserDetail `json:"users"`
}

// UserDetail represents detailed user information
type UserDetail struct {
	UserID     string        `json:"userId"`
	NodesCount int           `json:"nodesCount"`
	CreatedAt  time.Time     `json:"createdAt"`
	Nodes      []UserNodeInfo `json:"nodes"`
}

// UserNodeInfo represents a node where user was created
type UserNodeInfo struct {
	Endpoint    string    `json:"endpoint"`
	CountryCode string    `json:"countryCode"`
	CityName    string    `json:"cityName"`
	Inbound     string    `json:"inbound"`
	CreatedAt   time.Time `json:"createdAt"`
}

// NodesUsersResponse for GET /api/nodes/users
type NodesUsersResponse struct {
	Nodes []NodeUsersDetail `json:"nodes"`
}

// NodeUsersDetail represents users on a specific node
type NodeUsersDetail struct {
	Endpoint    string           `json:"endpoint"`
	CountryCode string           `json:"countryCode"`
	CityName    string           `json:"cityName"`
	InboundType string           `json:"inboundType"`
	Active      bool             `json:"active"`
	UsersCount  int              `json:"usersCount"`
	Users       []NodeUserInfo   `json:"users"`
}

// NodeUserInfo represents a user on a node
type NodeUserInfo struct {
	UserID    string    `json:"userId"`
	Inbound   string    `json:"inbound"`
	CreatedAt time.Time `json:"createdAt"`
}

// PendingDeletion represents a failed deletion attempt
type PendingDeletion struct {
	UserID       string    `json:"userId" db:"user_id"`
	Endpoint     string    `json:"endpoint" db:"endpoint"`
	Inbound      string    `json:"inbound" db:"inbound"`
	Attempts     int       `json:"attempts" db:"attempts"`
	LastAttempt  time.Time `json:"lastAttempt" db:"last_attempt"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	ErrorMessage string    `json:"errorMessage" db:"error_message"`
}

// PendingDeletionsResponse for GET /api/cleanup/pending
type PendingDeletionsResponse struct {
	Count            int               `json:"count"`
	PendingDeletions []PendingDeletion `json:"pendingDeletions"`
}

// CleanupResult for POST /api/cleanup
type CleanupResult struct {
	TotalAttempted int      `json:"totalAttempted"`
	Successful     int      `json:"successful"`
	Failed         int      `json:"failed"`
	StillPending   int      `json:"stillPending"`
	Errors         []string `json:"errors,omitempty"`
}

// DeletePendingRequest for DELETE /api/cleanup/pending
type DeletePendingRequest struct {
	UserID   string `json:"userId"`
	Endpoint string `json:"endpoint"`
	Inbound  string `json:"inbound,omitempty"` // Optional, if not provided will delete all for user+endpoint
}

