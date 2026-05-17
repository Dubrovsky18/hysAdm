package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Hysteria2Link struct {
	Server        string `json:"server"`
	Port          int    `json:"port"`
	Auth          string `json:"auth"`
	Transport     string `json:"transport"`
	TLS           *TLSConfig `json:"tls,omitempty"`
	Bandwidth     *Bandwidth `json:"bandwidth,omitempty"`
	QUIC          *QUICParams `json:"quic,omitempty"`
	Masquerade    string `json:"masquerade,omitempty"`
	Remark        string `json:"remark,omitempty"`
}

type TLSConfig struct {
	SNI           string `json:"sni"`
	Insecure      bool   `json:"insecure"`
}

type Bandwidth struct {
	Up   string `json:"up"`
	Down string `json:"down"`
}

type QUICParams struct {
	InitStreamRxWindow  int `json:"initStreamRxWindow,omitempty"`
	MaxStreamRxWindow   int `json:"maxStreamRxWindow,omitempty"`
	InitConnRxWindow    int `json:"initConnRxWindow,omitempty"`
	MaxConnRxWindow     int `json:"maxConnRxWindow,omitempty"`
	IdleTimeout         int `json:"idleTimeout,omitempty"`
}

func GenerateHysteria2Link(server, password, sni, remark string, port int, insecure bool) string {
	params := fmt.Sprintf("insecure=%d", 0)
	if insecure {
		params = "insecure=1"
	}

	if sni != "" {
		params += fmt.Sprintf("&sni=%s", sni)
	}

	return fmt.Sprintf("hysteria2://%s@%s:%d/?%s#%s",
		password, server, port, params, remark)
}

type SubscriptionSet struct {
	Version int                `json:"version"`
	SubName string             `json:"sub_name"`
	Nodes   []*Hysteria2Link  `json:"nodes"`
}

func GenerateSubscriptionResponse(links []string) (string, error) {
	var content string
	for _, link := range links {
		content += link + "\n"
	}
	return base64.StdEncoding.EncodeToString([]byte(content)), nil
}

// Deprecated: use GenerateClashConfig instead
func GenerateClashMetaConfig(links []string) (string, error) {
	type Proxy struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Server   string `json:"server"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		SNI      string `json:"sni,omitempty"`
		SkipCertVerify bool `json:"skip-cert-verify"`
	}

	var proxies []Proxy
	for i := range links {
		proxies = append(proxies, Proxy{
			Name:     fmt.Sprintf("Server-%d", i+1),
			Type:     "hysteria2",
			Server:   "example.com",
			Port:     443,
			Password: "password",
			SkipCertVerify: true,
		})
	}

	cfg := map[string]interface{}{
		"proxies": proxies,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ProxyInfo describes a Hysteria2 proxy for Clash Meta config
type ProxyInfo struct {
	Name     string
	Server   string
	Port     int
	Password string
	SNI      string
	Remark   string
}

// DomainRule maps a domain suffix to a target proxy server name
type DomainRule struct {
	Domain     string
	ServerName string
}

// ClashProxy is a single proxy entry in Clash Meta YAML
type ClashProxy struct {
	Name           string `yaml:"name"`
	Type           string `yaml:"type"`
	Server         string `yaml:"server"`
	Port           int    `yaml:"port"`
	Password       string `yaml:"password"`
	SNI            string `yaml:"sni,omitempty"`
	SkipCertVerify bool   `yaml:"skip-cert-verify"`
}

// ClashProxyGroup is a proxy group for routing
type ClashProxyGroup struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Proxies []string `yaml:"proxies"`
}

// ClashConfig is the root Clash Meta configuration structure
type ClashConfig struct {
	Port        int               `yaml:"port,omitempty"`
	Mode        string            `yaml:"mode,omitempty"`
	LogLevel    string            `yaml:"log-level,omitempty"`
	Proxies     []ClashProxy      `yaml:"proxies"`
	ProxyGroups []ClashProxyGroup `yaml:"proxy-groups"`
	Rules       []string          `yaml:"rules"`
}

// GenerateClashConfig builds a Clash Meta YAML config for domain-based routing.
func GenerateClashConfig(proxies []ProxyInfo, rules []DomainRule) (string, error) {
	clashProxies := make([]ClashProxy, 0, len(proxies))
	proxyNames := make([]string, 0, len(proxies))
	seen := make(map[string]bool)
	for _, p := range proxies {
		if seen[p.Name] {
			continue
		}
		seen[p.Name] = true
		clashProxies = append(clashProxies, ClashProxy{
			Name:           p.Name,
			Type:           "hysteria2",
			Server:         p.Server,
			Port:           p.Port,
			Password:       p.Password,
			SNI:            p.SNI,
			SkipCertVerify: true,
		})
		proxyNames = append(proxyNames, p.Name)
	}

	clashRules := make([]string, 0, len(rules)+1)
	for _, r := range rules {
		clashRules = append(clashRules, fmt.Sprintf("DOMAIN-SUFFIX,%s,%s", r.Domain, r.ServerName))
	}
	clashRules = append(clashRules, "MATCH,Proxy")

	proxyGroups := []ClashProxyGroup{
		{Name: "Proxy", Type: "select", Proxies: proxyNames},
	}

	cfg := ClashConfig{
		Port:        7890,
		Mode:        "rule",
		LogLevel:    "warning",
		Proxies:     clashProxies,
		ProxyGroups: proxyGroups,
		Rules:       clashRules,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal clash config: %w", err)
	}
	return string(data), nil
}
