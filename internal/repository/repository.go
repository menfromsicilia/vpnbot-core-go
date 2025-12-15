package repository

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"vpnbot-core-go/internal/models"
)

type Repository struct {
	db *sql.DB
}

func New(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &Repository{db: db}
	if err := repo.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return repo, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) runMigrations() error {
	migration, err := os.ReadFile("migrations/001_initial.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = r.db.Exec(string(migration))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// Server operations

func (r *Repository) GetServers(activeOnly bool) ([]models.Server, error) {
	query := "SELECT country_code, city_name, ext_name, endpoint, inbound_type, active, created_at FROM servers"
	if activeOnly {
		query += " WHERE active = 1"
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		var s models.Server
		if err := rows.Scan(&s.CountryCode, &s.CityName, &s.ExtName, &s.Endpoint, &s.InboundType, &s.Active, &s.CreatedAt); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}

	return servers, rows.Err()
}

func (r *Repository) CreateServers(servers []models.Server) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO servers (country_code, city_name, ext_name, endpoint, inbound_type, active)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range servers {
		_, err := stmt.Exec(s.CountryCode, s.CityName, s.ExtName, s.Endpoint, s.InboundType, s.Active)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) UpdateServers(servers []models.Server) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE servers 
		SET country_code = ?, city_name = ?, ext_name = ?, inbound_type = ?, active = ?
		WHERE endpoint = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range servers {
		_, err := stmt.Exec(s.CountryCode, s.CityName, s.ExtName, s.InboundType, s.Active, s.Endpoint)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) DeleteServers(endpoints []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM servers WHERE endpoint = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, endpoint := range endpoints {
		_, err := stmt.Exec(endpoint)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UserNode operations

func (r *Repository) SaveUserNodes(userID string, nodes []models.UserNode) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO user_nodes (user_id, endpoint, inbound)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, node := range nodes {
		_, err := stmt.Exec(userID, node.Endpoint, node.Inbound)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetUserNodes(userID string) ([]models.UserNode, error) {
	rows, err := r.db.Query(`
		SELECT user_id, endpoint, inbound, created_at 
		FROM user_nodes 
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []models.UserNode
	for rows.Next() {
		var n models.UserNode
		if err := rows.Scan(&n.UserID, &n.Endpoint, &n.Inbound, &n.CreatedAt); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}

	return nodes, rows.Err()
}

func (r *Repository) DeleteUserNodes(userID string) error {
	_, err := r.db.Exec("DELETE FROM user_nodes WHERE user_id = ?", userID)
	return err
}

func (r *Repository) DeleteUserNode(userID, endpoint string) error {
	_, err := r.db.Exec("DELETE FROM user_nodes WHERE user_id = ? AND endpoint = ?", userID, endpoint)
	return err
}

func (r *Repository) GetTotalUsersCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(DISTINCT user_id) FROM user_nodes").Scan(&count)
	return count, err
}

func (r *Repository) GetUsersCountByEndpoint(endpoint string) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(DISTINCT user_id) FROM user_nodes WHERE endpoint = ?", endpoint).Scan(&count)
	return count, err
}

// GetUsersCountByInbound returns count of users grouped by inbound type
func (r *Repository) GetUsersCountByInbound() (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT inbound, COUNT(*) as count 
		FROM user_nodes 
		GROUP BY inbound
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var inbound string
		var count int
		if err := rows.Scan(&inbound, &count); err != nil {
			return nil, err
		}
		counts[inbound] = count
	}

	return counts, rows.Err()
}

// NodeStats represents statistics for a single node
type NodeStats struct {
	Endpoint     string `json:"endpoint"`
	CountryCode  string `json:"countryCode"`
	CityName     string `json:"cityName"`
	ExtName      string `json:"extName,omitempty"`
	InboundType  string `json:"inboundType"`
	Active       bool   `json:"active"`
	UsersCount   int    `json:"usersCount"`
}

// GetNodeStats returns statistics for all nodes
func (r *Repository) GetNodeStats() ([]NodeStats, error) {
	query := `
		SELECT 
			s.endpoint,
			s.country_code,
			s.city_name,
			s.ext_name,
			s.inbound_type,
			s.active,
			COALESCE(COUNT(DISTINCT un.user_id), 0) as users_count
		FROM servers s
		LEFT JOIN user_nodes un ON s.endpoint = un.endpoint
		GROUP BY s.endpoint, s.country_code, s.city_name, s.ext_name, s.inbound_type, s.active
		ORDER BY s.active DESC, users_count DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []NodeStats
	for rows.Next() {
		var ns NodeStats
		if err := rows.Scan(
			&ns.Endpoint,
			&ns.CountryCode,
			&ns.CityName,
			&ns.ExtName,
			&ns.InboundType,
			&ns.Active,
			&ns.UsersCount,
		); err != nil {
			return nil, err
		}
		stats = append(stats, ns)
	}

	return stats, rows.Err()
}

// GetAllUsersWithDetails returns all users with their nodes information
func (r *Repository) GetAllUsersWithDetails() ([]models.UserDetail, error) {
	query := `
		SELECT 
			un.user_id,
			un.endpoint,
			s.country_code,
			s.city_name,
			un.inbound,
			un.created_at
		FROM user_nodes un
		LEFT JOIN servers s ON un.endpoint = s.endpoint
		ORDER BY un.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map to group by user_id
	usersMap := make(map[string]*models.UserDetail)
	var orderedUserIDs []string

	for rows.Next() {
		var userID, endpoint, countryCode, cityName, inbound string
		var createdAt time.Time

		if err := rows.Scan(&userID, &endpoint, &countryCode, &cityName, &inbound, &createdAt); err != nil {
			return nil, err
		}

		// If user not in map, create entry
		if _, exists := usersMap[userID]; !exists {
			usersMap[userID] = &models.UserDetail{
				UserID:     userID,
				NodesCount: 0,
				CreatedAt:  createdAt,
				Nodes:      []models.UserNodeInfo{},
			}
			orderedUserIDs = append(orderedUserIDs, userID)
		}

		// Add node info
		usersMap[userID].Nodes = append(usersMap[userID].Nodes, models.UserNodeInfo{
			Endpoint:    endpoint,
			CountryCode: countryCode,
			CityName:    cityName,
			Inbound:     inbound,
			CreatedAt:   createdAt,
		})
		usersMap[userID].NodesCount++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert map to slice maintaining order
	var users []models.UserDetail
	for _, userID := range orderedUserIDs {
		users = append(users, *usersMap[userID])
	}

	return users, nil
}

// GetNodesWithUsers returns all nodes with their users information
func (r *Repository) GetNodesWithUsers() ([]models.NodeUsersDetail, error) {
	query := `
		SELECT 
			s.endpoint,
			s.country_code,
			s.city_name,
			s.inbound_type,
			s.active,
			COALESCE(un.user_id, '') as user_id,
			COALESCE(un.inbound, '') as inbound,
			COALESCE(un.created_at, '') as created_at
		FROM servers s
		LEFT JOIN user_nodes un ON s.endpoint = un.endpoint
		ORDER BY s.country_code, s.city_name, un.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map to group by endpoint
	nodesMap := make(map[string]*models.NodeUsersDetail)
	var orderedEndpoints []string

	for rows.Next() {
		var endpoint, countryCode, cityName, inboundType string
		var active bool
		var userID, inbound string
		var createdAtStr string

		if err := rows.Scan(&endpoint, &countryCode, &cityName, &inboundType, &active, &userID, &inbound, &createdAtStr); err != nil {
			return nil, err
		}

		// If node not in map, create entry
		if _, exists := nodesMap[endpoint]; !exists {
			nodesMap[endpoint] = &models.NodeUsersDetail{
				Endpoint:    endpoint,
				CountryCode: countryCode,
				CityName:    cityName,
				InboundType: inboundType,
				Active:      active,
				UsersCount:  0,
				Users:       []models.NodeUserInfo{},
			}
			orderedEndpoints = append(orderedEndpoints, endpoint)
		}

		// Add user info if exists
		if userID != "" {
			createdAt, _ := time.Parse("2006-01-02 15:04:05", createdAtStr)
			nodesMap[endpoint].Users = append(nodesMap[endpoint].Users, models.NodeUserInfo{
				UserID:    userID,
				Inbound:   inbound,
				CreatedAt: createdAt,
			})
			nodesMap[endpoint].UsersCount++
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert map to slice maintaining order
	var nodes []models.NodeUsersDetail
	for _, endpoint := range orderedEndpoints {
		nodes = append(nodes, *nodesMap[endpoint])
	}

	return nodes, nil
}

// Pending Deletions operations

func (r *Repository) AddPendingDeletion(userID, endpoint, inbound, errorMsg string) error {
	_, err := r.db.Exec(`
		INSERT INTO pending_deletions (user_id, endpoint, inbound, error_message)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, endpoint, inbound) DO UPDATE SET
			attempts = attempts + 1,
			last_attempt = CURRENT_TIMESTAMP,
			error_message = ?
	`, userID, endpoint, inbound, errorMsg, errorMsg)
	return err
}

func (r *Repository) GetPendingDeletions() ([]models.PendingDeletion, error) {
	rows, err := r.db.Query(`
		SELECT user_id, endpoint, inbound, attempts, last_attempt, created_at, error_message
		FROM pending_deletions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deletions []models.PendingDeletion
	for rows.Next() {
		var pd models.PendingDeletion
		if err := rows.Scan(&pd.UserID, &pd.Endpoint, &pd.Inbound, &pd.Attempts, &pd.LastAttempt, &pd.CreatedAt, &pd.ErrorMessage); err != nil {
			return nil, err
		}
		deletions = append(deletions, pd)
	}

	return deletions, rows.Err()
}

func (r *Repository) RemovePendingDeletion(userID, endpoint, inbound string) error {
	_, err := r.db.Exec(`
		DELETE FROM pending_deletions 
		WHERE user_id = ? AND endpoint = ? AND inbound = ?
	`, userID, endpoint, inbound)
	return err
}

func (r *Repository) GetPendingDeletionsCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM pending_deletions").Scan(&count)
	return count, err
}

