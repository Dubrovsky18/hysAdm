package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"hysteria2-panel/internal/middleware"
	"hysteria2-panel/internal/models"
	"hysteria2-panel/internal/services"
	"hysteria2-panel/internal/subscription"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	userService         *services.UserService
	serverService       *services.ServerService
	keyService          *services.KeyService
	subscriptionService *services.SubscriptionService
	domainService       *services.DomainService
	trafficService      *services.TrafficService
	jwtSecret           string
	jwtTTL              int
}

func NewHandler(pool *pgxpool.Pool, jwtSecret string, jwtTTL int, panelDomain string) *Handler {
	return &Handler{
		userService:         services.NewUserService(pool),
		serverService:       services.NewServerService(pool),
		keyService:          services.NewKeyService(pool),
		subscriptionService: services.NewSubscriptionService(pool, panelDomain),
		domainService:       services.NewDomainService(pool),
		trafficService:      services.NewTrafficService(pool),
		jwtSecret:           jwtSecret,
		jwtTTL:              jwtTTL,
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		jsonError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Duration(h.jwtTTL) * time.Second).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		jsonError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, models.LoginResponse{Token: tokenString})
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.trafficService.GetStats(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, stats)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		jsonError(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Create(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, user)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.List(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, users)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Update(r.Context(), id, req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.userService.Delete(r.Context(), id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "deleted"})
}

func (h *Handler) CreateServer(w http.ResponseWriter, r *http.Request) {
	var req models.CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	server, err := h.serverService.Create(r.Context(), req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, server)
}

func (h *Handler) ListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.serverService.List(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, servers)
}

func (h *Handler) GetServer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid server id", http.StatusBadRequest)
		return
	}

	server, err := h.serverService.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, "server not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, server)
}

func (h *Handler) UpdateServer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid server id", http.StatusBadRequest)
		return
	}

	var req models.CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	server, err := h.serverService.Update(r.Context(), id, req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, server)
}

func (h *Handler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid server id", http.StatusBadRequest)
		return
	}

	if err := h.serverService.Delete(r.Context(), id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "deleted"})
}

func (h *Handler) ToggleServer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid server id", http.StatusBadRequest)
		return
	}

	var req struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.serverService.ToggleActive(r.Context(), id, req.Active); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "updated"})
}

func (h *Handler) CreateKey(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req models.CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	us, err := h.keyService.CreateKey(r.Context(), userID, req.ServerID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, us)
}

func (h *Handler) ListUserKeys(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	keys, err := h.keyService.GetUserKeys(r.Context(), userID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, keys)
}

func (h *Handler) RevokeKey(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("keyId"), 10, 64)
	if err != nil {
		jsonError(w, "invalid key id", http.StatusBadRequest)
		return
	}

	if err := h.keyService.RevokeKey(r.Context(), id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "key revoked"})
}

func (h *Handler) GenerateSubscription(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	sub, err := h.subscriptionService.GenerateSubscription(r.Context(), userID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, sub)
}

func (h *Handler) GetSubscriptionHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	subs, err := h.subscriptionService.GetSubscriptionHistory(r.Context(), userID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, subs)
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		jsonError(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetByUUID(r.Context(), uuid)
	if err != nil {
		jsonError(w, "subscription not found", http.StatusNotFound)
		return
	}

	domainRules, err := h.domainService.GetDomainsForUserServers(r.Context(), user.ID)
	if err != nil {
		jsonError(w, "subscription not found", http.StatusNotFound)
		return
	}

	if len(domainRules) > 0 {
		proxies, err := h.subscriptionService.GetUserServerDetails(r.Context(), user.ID)
		if err != nil {
			jsonError(w, "subscription not found", http.StatusNotFound)
			return
		}

		if len(proxies) == 0 {
			jsonError(w, "no servers assigned", http.StatusNotFound)
			return
		}

		clashCfg, err := subscription.GenerateClashConfig(proxies, domainRules)
		if err != nil {
			jsonError(w, "failed to generate config", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		w.Header().Set("Subscription-Userinfo", "upload=0; download=0; total=107374182400; expire=0")
		w.Write([]byte(clashCfg))
		return
	}

	links, err := h.subscriptionService.GetSubscriptionLinks(r.Context(), uuid)
	if err != nil {
		jsonError(w, "subscription not found", http.StatusNotFound)
		return
	}

	var content string
	for _, link := range links {
		content += link + "\n"
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Subscription-Userinfo", "upload=0; download=0; total=107374182400; expire=0")
	w.Write([]byte(content))
}

func (h *Handler) CreateDomain(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain   string `json:"domain"`
		ServerID int64  `json:"server_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	domain, err := h.domainService.Create(r.Context(), req.Domain, req.ServerID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, domain)
}

func (h *Handler) ListDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := h.domainService.List(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, domains)
}

func (h *Handler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid domain id", http.StatusBadRequest)
		return
	}

	if err := h.domainService.Delete(r.Context(), id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "deleted"})
}

func (h *Handler) GetUserTraffic(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	stats, err := h.trafficService.GetUserTraffic(r.Context(), userID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, stats)
}

func (h *Handler) GetAllUserTraffic(w http.ResponseWriter, r *http.Request) {
	stats, err := h.trafficService.GetAllUserTraffic(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, stats)
}

func (h *Handler) RecordTraffic(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   int64 `json:"user_id"`
		ServerID int64 `json:"server_id"`
		Upload   int64 `json:"upload"`
		Download int64 `json:"download"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.trafficService.RecordTraffic(r.Context(), req.UserID, req.ServerID, req.Upload, req.Download); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "recorded"})
}

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, user)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
