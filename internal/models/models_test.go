package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserJSON(t *testing.T) {
	u := User{
		ID:             1,
		UUID:           "550e8400-e29b-41d4-a716-446655440000",
		Username:       "testuser",
		Email:          "test@example.com",
		Status:         "active",
		TrafficUsed:    1024,
		TrafficLimit:   1073741824,
		SubscriptionURL: "hysteria2://panel.example.com/sub/uuid",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("failed to marshal User: %v", err)
	}

	var decoded User
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal User: %v", err)
	}

	if decoded.ID != u.ID {
		t.Errorf("ID: got %d, want %d", decoded.ID, u.ID)
	}
	if decoded.Username != u.Username {
		t.Errorf("Username: got %s, want %s", decoded.Username, u.Username)
	}
	if decoded.PasswordHash != "" {
		t.Error("PasswordHash should be hidden from JSON")
	}
}

func TestServerJSON(t *testing.T) {
	s := Server{
		ID:        1,
		Name:      "RU-Server",
		Address:   "ru.example.com",
		Port:      443,
		APIPort:   9443,
		APIKey:    "secret-key",
		Location:  "Russia",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal Server: %v", err)
	}

	var decoded Server
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Server: %v", err)
	}

	if decoded.Name != s.Name {
		t.Errorf("Name: got %s, want %s", decoded.Name, s.Name)
	}
	if !decoded.IsActive {
		t.Error("IsActive should be true")
	}
}

func TestDomainJSON(t *testing.T) {
	d := Domain{
		ID:        1,
		Domain:    "ru",
		ServerID:  2,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("failed to marshal Domain: %v", err)
	}

	var decoded Domain
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Domain: %v", err)
	}

	if decoded.Domain != d.Domain {
		t.Errorf("Domain: got %s, want %s", decoded.Domain, d.Domain)
	}
	if decoded.ServerID != d.ServerID {
		t.Errorf("ServerID: got %d, want %d", decoded.ServerID, d.ServerID)
	}
}

func TestCreateUserRequest(t *testing.T) {
	req := CreateUserRequest{
		Username:     "newuser",
		Password:     "secret123",
		Email:        "new@example.com",
		TrafficLimit: 1073741824,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal CreateUserRequest: %v", err)
	}

	var decoded CreateUserRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CreateUserRequest: %v", err)
	}

	if decoded.Username != req.Username {
		t.Errorf("Username: got %s, want %s", decoded.Username, req.Username)
	}
	if decoded.TrafficLimit != req.TrafficLimit {
		t.Errorf("TrafficLimit: got %d, want %d", decoded.TrafficLimit, req.TrafficLimit)
	}
}

func TestDashboardStats(t *testing.T) {
	stats := DashboardStats{
		TotalUsers:   10,
		ActiveUsers:  8,
		TotalServers: 3,
		TotalTraffic: 1000000,
		TodayTraffic: 50000,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("failed to marshal DashboardStats: %v", err)
	}

	var decoded DashboardStats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal DashboardStats: %v", err)
	}

	if decoded.TotalUsers != stats.TotalUsers {
		t.Errorf("TotalUsers: got %d, want %d", decoded.TotalUsers, stats.TotalUsers)
	}
}
