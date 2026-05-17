package subscription

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateHysteria2Link(t *testing.T) {
	tests := []struct {
		name     string
		server   string
		password string
		sni      string
		remark   string
		port     int
		insecure bool
		want     string
	}{
		{
			name:     "basic with sni",
			server:   "ru.example.com",
			password: "abc123",
			sni:      "panel.example.com",
			remark:   "RU-Server",
			port:     443,
			insecure: true,
			want:     "hysteria2://abc123@ru.example.com:443/?insecure=1&sni=panel.example.com#RU-Server",
		},
		{
			name:     "without sni",
			server:   "us.example.com",
			password: "key456",
			sni:      "",
			remark:   "US-Server",
			port:     8443,
			insecure: false,
			want:     "hysteria2://key456@us.example.com:8443/?insecure=0#US-Server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateHysteria2Link(tt.server, tt.password, tt.sni, tt.remark, tt.port, tt.insecure)
			if got != tt.want {
				t.Errorf("GenerateHysteria2Link() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateSubscriptionResponse(t *testing.T) {
	links := []string{
		"hysteria2://key1@server1:443/?insecure=1#Server1",
		"hysteria2://key2@server2:443/?insecure=1#Server2",
	}

	got, err := GenerateSubscriptionResponse(links)
	if err != nil {
		t.Fatalf("GenerateSubscriptionResponse() error = %v", err)
	}

	if got == "" {
		t.Error("GenerateSubscriptionResponse() returned empty string")
	}
}

func TestGenerateClashConfig_basic(t *testing.T) {
	proxies := []ProxyInfo{
		{Name: "RU-Server", Server: "ru.example.com", Port: 443, Password: "key1", SNI: "panel.example.com"},
		{Name: "AI-Server", Server: "ai.example.com", Port: 443, Password: "key2", SNI: "panel.example.com"},
	}

	rules := []DomainRule{
		{Domain: "ru", ServerName: "RU-Server"},
		{Domain: "ai", ServerName: "AI-Server"},
	}

	got, err := GenerateClashConfig(proxies, rules)
	if err != nil {
		t.Fatalf("GenerateClashConfig() error = %v", err)
	}

	var cfg ClashConfig
	if err := yaml.Unmarshal([]byte(got), &cfg); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	if len(cfg.Proxies) != 2 {
		t.Errorf("expected 2 proxies, got %d", len(cfg.Proxies))
	}
	if cfg.Proxies[0].Type != "hysteria2" {
		t.Errorf("expected proxy type hysteria2, got %s", cfg.Proxies[0].Type)
	}
	if len(cfg.ProxyGroups) != 1 {
		t.Errorf("expected 1 proxy group, got %d", len(cfg.ProxyGroups))
	}
	if cfg.ProxyGroups[0].Name != "Proxy" {
		t.Errorf("expected Proxy group, got %s", cfg.ProxyGroups[0].Name)
	}
	if len(cfg.Rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0] != "DOMAIN-SUFFIX,ru,RU-Server" {
		t.Errorf("unexpected rule: %s", cfg.Rules[0])
	}
	if cfg.Rules[2] != "MATCH,Proxy" {
		t.Errorf("expected MATCH rule, got %s", cfg.Rules[2])
	}
	if cfg.Mode != "rule" {
		t.Errorf("expected mode 'rule', got %s", cfg.Mode)
	}
}

func TestGenerateClashConfig_emptyRules(t *testing.T) {
	proxies := []ProxyInfo{
		{Name: "Main", Server: "main.example.com", Port: 443, Password: "key1", SNI: "panel.example.com"},
	}

	got, err := GenerateClashConfig(proxies, nil)
	if err != nil {
		t.Fatalf("GenerateClashConfig() error = %v", err)
	}

	var cfg ClashConfig
	if err := yaml.Unmarshal([]byte(got), &cfg); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	if len(cfg.Proxies) != 1 {
		t.Errorf("expected 1 proxy, got %d", len(cfg.Proxies))
	}
	if len(cfg.Rules) != 1 {
		t.Errorf("expected 1 rule (MATCH), got %d", len(cfg.Rules))
	}
	if cfg.Rules[0] != "MATCH,Proxy" {
		t.Errorf("expected MATCH rule, got %s", cfg.Rules[0])
	}
}

func TestGenerateClashConfig_dedup(t *testing.T) {
	proxies := []ProxyInfo{
		{Name: "Main", Server: "main.example.com", Port: 443, Password: "key1", SNI: "panel.example.com"},
		{Name: "Main", Server: "main.example.com", Port: 443, Password: "key1", SNI: "panel.example.com"},
	}

	got, err := GenerateClashConfig(proxies, nil)
	if err != nil {
		t.Fatalf("GenerateClashConfig() error = %v", err)
	}

	var cfg ClashConfig
	if err := yaml.Unmarshal([]byte(got), &cfg); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	if len(cfg.Proxies) != 1 {
		t.Errorf("expected 1 proxy after dedup, got %d", len(cfg.Proxies))
	}
}

func TestGenerateClashConfig_noProxies(t *testing.T) {
	got, err := GenerateClashConfig(nil, nil)
	if err != nil {
		t.Fatalf("GenerateClashConfig() error = %v", err)
	}

	var cfg ClashConfig
	if err := yaml.Unmarshal([]byte(got), &cfg); err != nil {
		t.Fatalf("invalid YAML output: %v", err)
	}

	if len(cfg.Proxies) != 0 {
		t.Errorf("expected 0 proxies, got %d", len(cfg.Proxies))
	}
	if len(cfg.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(cfg.Rules))
	}
}
