package models

import "time"

type User struct {
	ID             int64     `json:"id"`
	UUID           string    `json:"uuid"`
	Username       string    `json:"username"`
	PasswordHash   string    `json:"-"`
	Email          string    `json:"email"`
	Status         string    `json:"status"`
	TrafficUsed    int64     `json:"traffic_used"`
	TrafficLimit   int64     `json:"traffic_limit"`
	SubscriptionURL string   `json:"subscription_url"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Server struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Port      int       `json:"port"`
	APIPort   int       `json:"api_port"`
	APIKey    string    `json:"api_key"`
	Location  string    `json:"location"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type UserServer struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	ServerID      int64     `json:"server_id"`
	Port          int       `json:"port"`
	Key           string    `json:"key"`
	TrafficUsed   int64     `json:"traffic_used"`
	TrafficLimit  int64     `json:"traffic_limit"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type Subscription struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	UserServerID int64     `json:"user_server_id"`
	Link         string    `json:"link"`
	UUID         string    `json:"uuid"`
	TrafficUsed  int64     `json:"traffic_used"`
	TrafficLimit int64     `json:"traffic_limit"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type TrafficLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ServerID  int64     `json:"server_id"`
	Upload    int64     `json:"upload"`
	Download  int64     `json:"download"`
	CreatedAt time.Time `json:"created_at"`
}

type Domain struct {
	ID        int64     `json:"id"`
	Domain    string    `json:"domain"`
	ServerID  int64     `json:"server_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type HysteriaUserConfig struct {
	Password string `json:"password"`
}

type HysteriaConfig struct {
	Listen     string                        `json:"listen"`
	Protocol   string                        `json:"protocol"`
	Cert       string                        `json:"cert"`
	Key        string                        `json:"key"`
	Auth       AuthConfig                    `json:"auth"`
	Bandwidth  BandwidthConfig               `json:"bandwidth"`
	QUIC       QUICConfig                    `json:"quic"`
	Users      map[string]HysteriaUserConfig `json:"users,omitempty"`
}

type AuthConfig struct {
	Type      string          `json:"type"`
	UserPass  map[string]string `json:"userpass,omitempty"`
	HTTP      *HTTPAuthConfig `json:"http,omitempty"`
}

type HTTPAuthConfig struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}

type BandwidthConfig struct {
	Up   string `json:"up"`
	Down string `json:"down"`
}

type QUICConfig struct {
	InitStreamReceiveWindow     int `json:"initStreamReceiveWindow"`
	MaxStreamReceiveWindow      int `json:"maxStreamReceiveWindow"`
	InitConnReceiveWindow       int `json:"initConnReceiveWindow"`
	MaxConnReceiveWindow        int `json:"maxConnReceiveWindow"`
	MaxIdleTimeout              int `json:"maxIdleTimeout"`
}

// SubscriptionLink represents a hysteria2 subscription entry
type SubscriptionLink struct {
	Server     string `json:"server"`
	Port       int    `json:"port"`
	Password   string `json:"password"`
	SNI        string `json:"sni"`
	SkipCertVerify bool `json:"insecure"`
	Up         string `json:"up"`
	Down       string `json:"down"`
	Protocol   string `json:"protocol"`
	UUID       string `json:"uuid"`
	Remark     string `json:"remark"`
}

// API types
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type CreateUserRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Email        string `json:"email"`
	TrafficLimit int64  `json:"traffic_limit"`
}

type UpdateUserRequest struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Email        string `json:"email,omitempty"`
	Status       string `json:"status,omitempty"`
	TrafficLimit int64  `json:"traffic_limit,omitempty"`
}

type CreateServerRequest struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Port     int    `json:"port"`
	APIPort  int    `json:"api_port"`
	APIKey   string `json:"api_key"`
	Location string `json:"location"`
}

type CreateKeyRequest struct {
	ServerID int64 `json:"server_id"`
}

type TrafficStats struct {
	UserID      int64 `json:"user_id"`
	Username    string `json:"username"`
	Upload      int64 `json:"upload"`
	Download    int64 `json:"download"`
	TotalUsed   int64 `json:"total_used"`
	TrafficLimit int64 `json:"traffic_limit"`
}

type DashboardStats struct {
	TotalUsers      int64 `json:"total_users"`
	ActiveUsers     int64 `json:"active_users"`
	TotalServers    int   `json:"total_servers"`
	TotalTraffic    int64 `json:"total_traffic"`
	TodayTraffic    int64 `json:"today_traffic"`
}
