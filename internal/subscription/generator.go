package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	for i, link := range links {
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
