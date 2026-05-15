package services

import (
	"context"
	"fmt"
	"time"

	"hysteria2-panel/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	pool *pgxpool.Pool
}

func NewUserService(pool *pgxpool.Pool) *UserService {
	return &UserService{pool: pool}
}

func (s *UserService) Create(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	uuid := generateUUID()
	user := &models.User{}

	err = s.pool.QueryRow(ctx,
		`INSERT INTO users (uuid, username, password_hash, email, traffic_limit) 
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, uuid, username, email, status, traffic_used, traffic_limit, subscription_url, created_at, updated_at`,
		uuid, req.Username, string(hash), req.Email, req.TrafficLimit,
	).Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.Status,
		&user.TrafficUsed, &user.TrafficLimit, &user.SubscriptionURL, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, uuid, username, email, status, traffic_used, traffic_limit, subscription_url, created_at, updated_at 
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.Status,
		&user.TrafficUsed, &user.TrafficLimit, &user.SubscriptionURL, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *UserService) GetByUUID(ctx context.Context, uuid string) (*models.User, error) {
	user := &models.User{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, uuid, username, email, status, traffic_used, traffic_limit, subscription_url, created_at, updated_at 
		 FROM users WHERE uuid = $1`, uuid,
	).Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.Status,
		&user.TrafficUsed, &user.TrafficLimit, &user.SubscriptionURL, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("get user by uuid: %w", err)
	}
	return user, nil
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, uuid, username, email, status, traffic_used, traffic_limit, subscription_url, created_at, updated_at 
		 FROM users WHERE username = $1`, username,
	).Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.Status,
		&user.TrafficUsed, &user.TrafficLimit, &user.SubscriptionURL, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (s *UserService) List(ctx context.Context) ([]*models.User, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, uuid, username, email, status, traffic_used, traffic_limit, subscription_url, created_at, updated_at 
		 FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.UUID, &user.Username, &user.Email, &user.Status,
			&user.TrafficUsed, &user.TrafficLimit, &user.SubscriptionURL, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *UserService) Update(ctx context.Context, id int64, req models.UpdateUserRequest) (*models.User, error) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		_, err = s.pool.Exec(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, string(hash), id)
		if err != nil {
			return nil, err
		}
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Status != "" {
		user.Status = req.Status
	}
	if req.TrafficLimit > 0 {
		user.TrafficLimit = req.TrafficLimit
	}

	_, err = s.pool.Exec(ctx,
		`UPDATE users SET username=$1, email=$2, status=$3, traffic_limit=$4, updated_at=NOW() WHERE id=$5`,
		user.Username, user.Email, user.Status, user.TrafficLimit, id)
	if err != nil {
		return nil, err
	}

	user.UpdatedAt = time.Now()
	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (*models.User, error) {
	user := &models.User{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, uuid, username, password_hash, email, status, traffic_used, traffic_limit, subscription_url, created_at, updated_at 
		 FROM users WHERE username = $1`, username,
	).Scan(&user.ID, &user.UUID, &user.Username, &user.PasswordHash, &user.Email, &user.Status,
		&user.TrafficUsed, &user.TrafficLimit, &user.SubscriptionURL, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

func (s *UserService) UpdateTraffic(ctx context.Context, userID int64, trafficUsed int64) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET traffic_used = traffic_used + $1, updated_at = NOW() WHERE id = $2`,
		trafficUsed, userID)
	return err
}
