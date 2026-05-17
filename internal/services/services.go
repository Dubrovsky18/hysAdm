package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"hysteria2-panel/internal/models"
	"hysteria2-panel/internal/subscription"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerService struct {
	pool *pgxpool.Pool
}

func NewServerService(pool *pgxpool.Pool) *ServerService {
	return &ServerService{pool: pool}
}

func (s *ServerService) Create(ctx context.Context, req models.CreateServerRequest) (*models.Server, error) {
	server := &models.Server{}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO servers (name, address, port, api_port, api_key, location)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, name, address, port, api_port, api_key, location, is_active, created_at`,
		req.Name, req.Address, req.Port, req.APIPort, req.APIKey, req.Location,
	).Scan(&server.ID, &server.Name, &server.Address, &server.Port, &server.APIPort,
		&server.APIKey, &server.Location, &server.IsActive, &server.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("create server: %w", err)
	}
	return server, nil
}

func (s *ServerService) GetByID(ctx context.Context, id int64) (*models.Server, error) {
	server := &models.Server{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, address, port, api_port, api_key, location, is_active, created_at 
		 FROM servers WHERE id = $1`, id,
	).Scan(&server.ID, &server.Name, &server.Address, &server.Port, &server.APIPort,
		&server.APIKey, &server.Location, &server.IsActive, &server.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("get server: %w", err)
	}
	return server, nil
}

func (s *ServerService) List(ctx context.Context) ([]*models.Server, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, address, port, api_port, api_key, location, is_active, created_at 
		 FROM servers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := make([]*models.Server, 0)
	for rows.Next() {
		server := &models.Server{}
		if err := rows.Scan(&server.ID, &server.Name, &server.Address, &server.Port, &server.APIPort,
			&server.APIKey, &server.Location, &server.IsActive, &server.CreatedAt); err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}
	return servers, nil
}

func (s *ServerService) Update(ctx context.Context, id int64, req models.CreateServerRequest) (*models.Server, error) {
	_, err := s.pool.Exec(ctx,
		`UPDATE servers SET name=$1, address=$2, port=$3, api_port=$4, api_key=$5, location=$6 WHERE id=$7`,
		req.Name, req.Address, req.Port, req.APIPort, req.APIKey, req.Location, id)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *ServerService) Delete(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM servers WHERE id = $1`, id)
	return err
}

func (s *ServerService) ToggleActive(ctx context.Context, id int64, active bool) error {
	_, err := s.pool.Exec(ctx, `UPDATE servers SET is_active = $1 WHERE id = $2`, active, id)
	return err
}

type KeyService struct {
	pool *pgxpool.Pool
}

func NewKeyService(pool *pgxpool.Pool) *KeyService {
	return &KeyService{pool: pool}
}

func (s *KeyService) CreateKey(ctx context.Context, userID, serverID int64) (*models.UserServer, error) {
	key := generateKey()
	us := &models.UserServer{}

	err := s.pool.QueryRow(ctx,
		`INSERT INTO user_servers (user_id, server_id, key, traffic_limit)
		 VALUES ($1, $2, $3, (SELECT traffic_limit FROM users WHERE id = $1))
		 ON CONFLICT (user_id, server_id) DO UPDATE SET key = $3, is_active = true
		 RETURNING id, user_id, server_id, port, key, traffic_used, traffic_limit, is_active, created_at`,
		userID, serverID, key,
	).Scan(&us.ID, &us.UserID, &us.ServerID, &us.Port, &us.Key,
		&us.TrafficUsed, &us.TrafficLimit, &us.IsActive, &us.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("create key: %w", err)
	}

	return us, nil
}

func (s *KeyService) GetUserKeys(ctx context.Context, userID int64) ([]*models.UserServer, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT us.id, us.user_id, us.server_id, us.port, us.key, us.traffic_used, us.traffic_limit, us.is_active, us.created_at
		 FROM user_servers us WHERE us.user_id = $1 ORDER BY us.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]*models.UserServer, 0)
	for rows.Next() {
		us := &models.UserServer{}
		if err := rows.Scan(&us.ID, &us.UserID, &us.ServerID, &us.Port, &us.Key,
			&us.TrafficUsed, &us.TrafficLimit, &us.IsActive, &us.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, us)
	}
	return keys, nil
}

func (s *KeyService) RevokeKey(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `UPDATE user_servers SET is_active = false WHERE id = $1`, id)
	return err
}

type SubscriptionService struct {
	pool          *pgxpool.Pool
	panelDomain   string
}

func NewSubscriptionService(pool *pgxpool.Pool, panelDomain string) *SubscriptionService {
	return &SubscriptionService{pool: pool, panelDomain: panelDomain}
}

func (s *SubscriptionService) GenerateSubscription(ctx context.Context, userID int64) (*models.Subscription, error) {
	user, err := getUserByID(ctx, s.pool, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx,
		`SELECT us.id, s.address, s.port, us.key, us.traffic_limit
		 FROM user_servers us JOIN servers s ON us.server_id = s.id
		 WHERE us.user_id = $1 AND us.is_active = true AND s.is_active = true`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]string, 0)
	for rows.Next() {
		var usID int64
		var address string
		var port int
		var key string
		var limit int64

		if err := rows.Scan(&usID, &address, &port, &key, &limit); err != nil {
			return nil, err
		}

		link := fmt.Sprintf("hysteria2://%s@%s:%d/?insecure=1&sni=%s&obfs=none#%s-%s",
			key, address, port, s.panelDomain, s.panelDomain, address)
		links = append(links, link)

		subUUID := generateUUID()
		_, err = s.pool.Exec(ctx,
			`INSERT INTO subscriptions (user_id, user_server_id, link, uuid, traffic_limit, expires_at)
			 VALUES ($1, $2, $3, $4, $5, NOW() + INTERVAL '30 days')`,
			userID, usID, link, subUUID, limit)
		if err != nil {
			return nil, err
		}
	}

	subURL := fmt.Sprintf("hysteria2://%s/sub/%s", s.panelDomain, user.UUID)
	_, err = s.pool.Exec(ctx, `UPDATE users SET subscription_url = $1 WHERE id = $2`, subURL, userID)

	return &models.Subscription{Link: subURL}, err
}

func (s *SubscriptionService) GetUserServerDetails(ctx context.Context, userID int64) ([]subscription.ProxyInfo, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT s.name, s.address, s.port, us.key
		 FROM user_servers us
		 JOIN servers s ON us.server_id = s.id
		 WHERE us.user_id = $1 AND us.is_active = true AND s.is_active = true`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	proxies := make([]subscription.ProxyInfo, 0)
	for rows.Next() {
		var name, address, key string
		var port int
		if err := rows.Scan(&name, &address, &port, &key); err != nil {
			return nil, err
		}
		proxies = append(proxies, subscription.ProxyInfo{
			Name:     name,
			Server:   address,
			Port:     port,
			Password: key,
			SNI:      s.panelDomain,
			Remark:   fmt.Sprintf("%s-%s", s.panelDomain, name),
		})
	}
	return proxies, nil
}

func (s *DomainService) GetDomainsForUserServers(ctx context.Context, userID int64) ([]subscription.DomainRule, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT d.domain, s.name as server_name
		 FROM domains d
		 JOIN servers s ON d.server_id = s.id
		 JOIN user_servers us ON us.server_id = s.id AND us.user_id = $1
		 WHERE d.is_active = true AND s.is_active = true AND us.is_active = true`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]subscription.DomainRule, 0)
	for rows.Next() {
		var domain, serverName string
		if err := rows.Scan(&domain, &serverName); err != nil {
			return nil, err
		}
		rules = append(rules, subscription.DomainRule{
			Domain:     domain,
			ServerName: serverName,
		})
	}
	return rules, nil
}

func (s *SubscriptionService) GetSubscriptionLinks(ctx context.Context, userUUID string) ([]string, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT sub.link FROM subscriptions sub
		 JOIN users u ON sub.user_id = u.id
		 WHERE u.uuid = $1 AND sub.is_active = true
		 ORDER BY sub.created_at DESC`, userUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]string, 0)
	for rows.Next() {
		var link string
		if err := rows.Scan(&link); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	if len(links) == 0 {
		return nil, fmt.Errorf("no active subscriptions found")
	}

	return links, nil
}

func (s *SubscriptionService) GetSubscriptionHistory(ctx context.Context, userID int64) ([]*models.Subscription, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, user_server_id, link, uuid, traffic_used, traffic_limit, expires_at, is_active, created_at
		 FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := make([]*models.Subscription, 0)
	for rows.Next() {
		sub := &models.Subscription{}
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.UserServerID, &sub.Link, &sub.UUID,
			&sub.TrafficUsed, &sub.TrafficLimit, &sub.ExpiresAt, &sub.IsActive, &sub.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func getUserByID(ctx context.Context, pool *pgxpool.Pool, id int64) (*models.User, error) {
	user := &models.User{}
	err := pool.QueryRow(ctx,
		`SELECT id, uuid, username, email, status, traffic_used, traffic_limit, subscription_url, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.Status,
		&user.TrafficUsed, &user.TrafficLimit, &user.SubscriptionURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

type DomainService struct {
	pool *pgxpool.Pool
}

func NewDomainService(pool *pgxpool.Pool) *DomainService {
	return &DomainService{pool: pool}
}

func (s *DomainService) Create(ctx context.Context, domain string, serverID int64) (*models.Domain, error) {
	d := &models.Domain{}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO domains (domain, server_id) VALUES ($1, $2)
		 RETURNING id, domain, server_id, is_active, created_at`,
		domain, serverID,
	).Scan(&d.ID, &d.Domain, &d.ServerID, &d.IsActive, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create domain: %w", err)
	}
	return d, nil
}

func (s *DomainService) List(ctx context.Context) ([]*models.Domain, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, domain, server_id, is_active, created_at FROM domains ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	domains := make([]*models.Domain, 0)
	for rows.Next() {
		d := &models.Domain{}
		if err := rows.Scan(&d.ID, &d.Domain, &d.ServerID, &d.IsActive, &d.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, nil
}

func (s *DomainService) Delete(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM domains WHERE id = $1`, id)
	return err
}

type TrafficService struct {
	pool *pgxpool.Pool
}

func NewTrafficService(pool *pgxpool.Pool) *TrafficService {
	return &TrafficService{pool: pool}
}

func (s *TrafficService) RecordTraffic(ctx context.Context, userID, serverID, upload, download int64) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO traffic_logs (user_id, server_id, upload, download) VALUES ($1, $2, $3, $4)`,
		userID, serverID, upload, download)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx,
		`UPDATE users SET traffic_used = traffic_used + $1 + $2, updated_at = NOW() WHERE id = $3`,
		upload, download, userID)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx,
		`UPDATE user_servers SET traffic_used = traffic_used + $1 + $2 WHERE user_id = $3 AND server_id = $4`,
		upload, download, userID, serverID)
	return err
}

func (s *TrafficService) GetStats(ctx context.Context) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status IN ('active', 'admin')`).Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, err
	}

	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM servers WHERE is_active = true`).Scan(&stats.TotalServers)
	if err != nil {
		return nil, err
	}

	err = s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(upload + download), 0) FROM traffic_logs`).Scan(&stats.TotalTraffic)
	if err != nil {
		return nil, err
	}

	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(upload + download), 0) FROM traffic_logs WHERE created_at >= CURRENT_DATE`,
	).Scan(&stats.TodayTraffic)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *TrafficService) GetUserTraffic(ctx context.Context, userID int64) (*models.TrafficStats, error) {
	stats := &models.TrafficStats{UserID: userID}
	err := s.pool.QueryRow(ctx,
		`SELECT username, traffic_used, traffic_limit FROM users WHERE id = $1`, userID,
	).Scan(&stats.Username, &stats.TotalUsed, &stats.TrafficLimit)
	if err != nil {
		return nil, err
	}

	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(upload), 0), COALESCE(SUM(download), 0) FROM traffic_logs WHERE user_id = $1`,
		userID,
	).Scan(&stats.Upload, &stats.Download)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *TrafficService) GetAllUserTraffic(ctx context.Context) ([]*models.TrafficStats, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT u.id, u.username,
			COALESCE(SUM(t.upload), 0) as upload,
			COALESCE(SUM(t.download), 0) as download,
			u.traffic_used, u.traffic_limit
		 FROM users u
		 LEFT JOIN traffic_logs t ON u.id = t.user_id
		 GROUP BY u.id, u.username, u.traffic_used, u.traffic_limit
		 ORDER BY u.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]*models.TrafficStats, 0)
	for rows.Next() {
		s := &models.TrafficStats{}
		if err := rows.Scan(&s.UserID, &s.Username, &s.Upload, &s.Download, &s.TotalUsed, &s.TrafficLimit); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// Utility functions
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func generateKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
