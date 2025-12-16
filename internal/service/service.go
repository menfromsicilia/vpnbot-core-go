package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"vpnbot-core-go/internal/models"
	"vpnbot-core-go/internal/nodeclient"
	"vpnbot-core-go/internal/repository"
)

type Service struct {
	repo       *repository.Repository
	nodeClient *nodeclient.Client
	logger     *slog.Logger
	reqTimeout time.Duration
}

func New(repo *repository.Repository, nodeClient *nodeclient.Client, logger *slog.Logger, reqTimeout time.Duration) *Service {
	return &Service{
		repo:       repo,
		nodeClient: nodeClient,
		logger:     logger,
		reqTimeout: reqTimeout,
	}
}

type nodeResult struct {
	server models.Server
	config string
	err    error
}

// CreateUser creates a user on all active nodes in parallel
func (s *Service) CreateUser(ctx context.Context, existingUUID, targetEndpoint string) (*models.CreateUserResponse, error) {
	// Determine user ID
	var userID string
	if existingUUID != "" {
		userID = existingUUID
	} else {
		userID = uuid.New().String()
	}

	// Get servers to create on
	var servers []models.Server
	var err error
	
	if targetEndpoint != "" {
		// Create on specific node only
		server, err := s.repo.GetServerByEndpoint(targetEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to get server: %w", err)
		}
		if !server.Active {
			return nil, fmt.Errorf("server is not active")
		}
		servers = []models.Server{*server}
	} else {
		// Create on all active nodes
		servers, err = s.repo.GetServers(true)
		if err != nil {
			return nil, fmt.Errorf("failed to get servers: %w", err)
		}
	}

	if len(servers) == 0 {
		return nil, fmt.Errorf("no active servers available")
	}

	// First, delete user from all nodes (cleanup)
	s.deleteUserFromAllNodes(ctx, userID, servers)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.reqTimeout)
	defer cancel()

	// Create user on all nodes in parallel
	results := s.createUserOnNodes(ctx, userID, servers)

	// Process results
	var configs []models.ConfigItem
	var userNodes []models.UserNode
	var successCount, failCount int

	for _, result := range results {
		if result.err != nil {
			failCount++
			s.logger.Warn("node failed",
				slog.String("endpoint", result.server.Endpoint),
				slog.String("error", result.err.Error()),
			)
			continue
		}

		successCount++
		configs = append(configs, models.ConfigItem{
			CountryCode: result.server.CountryCode,
			IP:          result.server.Endpoint,
			Config:      result.config,
		})

		userNodes = append(userNodes, models.UserNode{
			UserID:   userID,
			Endpoint: result.server.Endpoint,
			Inbound:  result.server.InboundType,
		})
	}

	// Check if we have at least one successful config
	if successCount == 0 {
		return nil, fmt.Errorf("all nodes failed: %d/%d nodes unavailable", failCount, len(servers))
	}

	// Save user_nodes to database
	if err := s.repo.SaveUserNodes(userID, userNodes); err != nil {
		s.logger.Error("failed to save user nodes", slog.String("error", err.Error()))
		// Don't fail the request, configs are already created
	}

	s.logger.Info("user created",
		slog.String("user_id", userID),
		slog.Int("successful", successCount),
		slog.Int("failed", failCount),
	)

	return &models.CreateUserResponse{
		UUID:    userID,
		Configs: configs,
	}, nil
}

func (s *Service) createUserOnNodes(ctx context.Context, userID string, servers []models.Server) []nodeResult {
	results := make(chan nodeResult, len(servers))
	var wg sync.WaitGroup

	for _, server := range servers {
		wg.Add(1)
		go func(srv models.Server) {
			defer wg.Done()

			// Check context before starting
			select {
			case <-ctx.Done():
				results <- nodeResult{server: srv, err: ctx.Err()}
				return
			default:
			}

			xrayResp, err := s.nodeClient.CreateUser(srv.Endpoint, srv.InboundType, userID)
			if err != nil {
				results <- nodeResult{server: srv, err: err}
				return
			}

			// Build config string based on inbound type
			configStr, err := s.buildConfigString(srv, xrayResp)
			if err != nil {
				results <- nodeResult{server: srv, err: err}
				return
			}

			results <- nodeResult{server: srv, config: configStr, err: nil}
		}(server)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(results)

	// Collect results
	var allResults []nodeResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

func (s *Service) buildConfigString(server models.Server, xrayResp *models.XrayUserResponse) (string, error) {
	connCfg := xrayResp.ConnectionConfig

	switch server.InboundType {
	case "trojan":
		return s.buildTrojanConfig(server.Endpoint, xrayResp.ID, connCfg)
	case "vless":
		return s.buildVlessConfig(server.Endpoint, xrayResp.ID, connCfg)
	default:
		return "", fmt.Errorf("unsupported inbound type: %s", server.InboundType)
	}
}

func (s *Service) buildTrojanConfig(endpoint, userID string, connCfg map[string]interface{}) (string, error) {
	// Extract password
	password, ok := connCfg["password"].(string)
	if !ok {
		return "", fmt.Errorf("password not found in connection config")
	}

	// Extract tcp config with reality
	tcpCfg, ok := connCfg["tcp"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("tcp config not found")
	}

	port, ok := tcpCfg["port"].(float64)
	if !ok {
		return "", fmt.Errorf("port not found in tcp config")
	}

	realityCfg, ok := tcpCfg["reality"].(map[string]interface{})
	if !ok {
		// No reality, simple trojan
		return fmt.Sprintf("trojan://%s@%s:%.0f?type=tcp&security=none#Config:%s",
			password, endpoint, port, userID), nil
	}

	// Build trojan with reality
	fp := getString(realityCfg, "fingerprint", "chrome")
	sni := getString(realityCfg, "serverName", "")
	pbk := getString(realityCfg, "public_key", "")
	spx := getString(realityCfg, "spiderX", "")
	sid := getString(realityCfg, "shortId", "")

	return fmt.Sprintf("trojan://%s@%s:%.0f?type=tcp&security=reality&fp=%s&sni=%s&pbk=%s&sid=%s&spx=%s#Config:%s",
		password, endpoint, port, fp, sni, pbk, sid, spx, userID), nil
}

func (s *Service) buildVlessConfig(endpoint, userID string, connCfg map[string]interface{}) (string, error) {
	// Try to get tcp_reality first (preferred)
	if tcpRealityCfg, ok := connCfg["tcp_reality"].(map[string]interface{}); ok {
		port, _ := tcpRealityCfg["port"].(float64)
		realityCfg, _ := tcpRealityCfg["reality"].(map[string]interface{})

		flow := getString(connCfg, "flow_reality", "xtls-rprx-vision")
		fp := getString(realityCfg, "fingerprint", "chrome")
		sni := getString(realityCfg, "serverName", "")
		pbk := getString(realityCfg, "public_key", "")
		spx := getString(realityCfg, "spiderX", "")
		sid := getString(realityCfg, "shortId", "")

		return fmt.Sprintf("vless://%s@%s:%.0f?flow=%s&type=tcp&security=reality&fp=%s&sni=%s&pbk=%s&sid=%s&spx=%s#XTLS-Reality:%s",
			userID, endpoint, port, flow, fp, sni, pbk, sid, spx, userID), nil
	}

	// Fallback to simple TCP
	if tcpCfg, ok := connCfg["tcp"].(map[string]interface{}); ok {
		port, _ := tcpCfg["port"].(float64)
		return fmt.Sprintf("vless://%s@%s:%.0f?type=tcp&security=none#TCP:%s",
			userID, endpoint, port, userID), nil
	}

	return "", fmt.Errorf("no valid vless config found")
}

func getString(m map[string]interface{}, key, defaultVal string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultVal
}

// DeleteUser deletes a user from all nodes where it was created
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	// Get nodes where user was created
	userNodes, err := s.repo.GetUserNodes(userID)
	if err != nil {
		return fmt.Errorf("failed to get user nodes: %w", err)
	}

	if len(userNodes) == 0 {
		// User not found, but that's OK (idempotent)
		s.logger.Info("user not found for deletion", slog.String("user_id", userID))
		return nil
	}

	// Delete from all nodes in parallel
	ctx, cancel := context.WithTimeout(ctx, s.reqTimeout)
	defer cancel()

	var wg sync.WaitGroup
	var successCount, failCount int
	var mu sync.Mutex

	for _, node := range userNodes {
		wg.Add(1)
		go func(n models.UserNode) {
			defer wg.Done()

			err := s.nodeClient.DeleteUser(n.Endpoint, userID)
			mu.Lock()
			if err != nil {
				failCount++
				s.logger.Warn("failed to delete user from node",
					slog.String("endpoint", n.Endpoint),
					slog.String("error", err.Error()),
				)
				// Record failed deletion for later cleanup
				if dbErr := s.repo.AddPendingDeletion(userID, n.Endpoint, n.Inbound, err.Error()); dbErr != nil {
					s.logger.Error("failed to record pending deletion",
						slog.String("error", dbErr.Error()),
					)
				}
			} else {
				successCount++
			}
			mu.Unlock()
		}(node)
	}

	wg.Wait()

	// Delete from database regardless of node results (best-effort)
	if err := s.repo.DeleteUserNodes(userID); err != nil {
		s.logger.Error("failed to delete user nodes from DB", slog.String("error", err.Error()))
	}

	s.logger.Info("user deleted",
		slog.String("user_id", userID),
		slog.Int("successful", successCount),
		slog.Int("failed", failCount),
	)

	return nil
}

// DeleteUserFromNode deletes a user from a specific node (backward compatibility)
func (s *Service) DeleteUserFromNode(ctx context.Context, userID, endpoint string) error {
	ctx, cancel := context.WithTimeout(ctx, s.reqTimeout)
	defer cancel()

	// Get node info to know inbound type (for tracking)
	nodes, err := s.repo.GetUserNodes(userID)
	var targetInbound string
	if err == nil {
		for _, node := range nodes {
			if node.Endpoint == endpoint {
				targetInbound = node.Inbound
				break
			}
		}
	}

	// Delete from the specific node
	if err := s.nodeClient.DeleteUser(endpoint, userID); err != nil {
		s.logger.Error("failed to delete user from node",
			slog.String("endpoint", endpoint),
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		
		// Record failed deletion if we know the inbound
		if targetInbound != "" {
			if dbErr := s.repo.AddPendingDeletion(userID, endpoint, targetInbound, err.Error()); dbErr != nil {
				s.logger.Error("failed to record pending deletion", slog.String("error", dbErr.Error()))
			}
		}
		
		return fmt.Errorf("failed to delete user from node: %w", err)
	}

	// Note: We don't remove from user_nodes table here because:
	// 1. Old users might not be in the table yet
	// 2. This is for backward compatibility with gradual migration

	s.logger.Info("user deleted from specific node",
		slog.String("user_id", userID),
		slog.String("endpoint", endpoint),
	)

	return nil
}

func (s *Service) deleteUserFromAllNodes(ctx context.Context, userID string, servers []models.Server) {
	var wg sync.WaitGroup
	for _, server := range servers {
		wg.Add(1)
		go func(endpoint string) {
			defer wg.Done()
			_ = s.nodeClient.DeleteUser(endpoint, userID)
		}(server.Endpoint)
	}
	wg.Wait()
}

// GetStats returns comprehensive statistics
func (s *Service) GetStats(ctx context.Context) (*models.StatsResponse, error) {
	// Get total users count
	totalUsers, err := s.repo.GetTotalUsersCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get total users count: %w", err)
	}

	// Get node statistics
	nodeStats, err := s.repo.GetNodeStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get node stats: %w", err)
	}

	// Convert to response format
	nodes := make([]models.NodeStatsItem, len(nodeStats))
	for i, ns := range nodeStats {
		nodes[i] = models.NodeStatsItem{
			Endpoint:    ns.Endpoint,
			CountryCode: ns.CountryCode,
			CityName:    ns.CityName,
			ExtName:     ns.ExtName,
			InboundType: ns.InboundType,
			Active:      ns.Active,
			UsersCount:  ns.UsersCount,
		}
	}

	// Get counts by protocol
	byProtocol, err := s.repo.GetUsersCountByInbound()
	if err != nil {
		return nil, fmt.Errorf("failed to get counts by inbound: %w", err)
	}

	return &models.StatsResponse{
		TotalUsers: totalUsers,
		Nodes:      nodes,
		ByProtocol: byProtocol,
	}, nil
}

// GetUsersCount returns total users count
func (s *Service) GetUsersCount(ctx context.Context) (int, error) {
	return s.repo.GetTotalUsersCount()
}

// GetAllUsers returns all users with their nodes information
func (s *Service) GetAllUsers(ctx context.Context) (*models.UserListResponse, error) {
	users, err := s.repo.GetAllUsersWithDetails()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return &models.UserListResponse{
		Users: users,
	}, nil
}

// GetNodesWithUsers returns all nodes with their users information
func (s *Service) GetNodesWithUsers(ctx context.Context) (*models.NodesUsersResponse, error) {
	nodes, err := s.repo.GetNodesWithUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes with users: %w", err)
	}

	return &models.NodesUsersResponse{
		Nodes: nodes,
	}, nil
}

// GetPendingDeletions returns list of failed deletion attempts
func (s *Service) GetPendingDeletions() ([]models.PendingDeletion, error) {
	return s.repo.GetPendingDeletions()
}

// DeletePendingDeletion manually removes a specific pending deletion record
func (s *Service) DeletePendingDeletion(userID, endpoint, inbound string) error {
	if userID == "" || endpoint == "" {
		return fmt.Errorf("userId and endpoint are required")
	}
	
	// If inbound is not specified, we need to get all pending deletions for this user+endpoint
	if inbound == "" {
		pendingDeletions, err := s.repo.GetPendingDeletions()
		if err != nil {
			return fmt.Errorf("failed to get pending deletions: %w", err)
		}
		
		// Delete all matching user+endpoint combinations
		deleted := 0
		for _, pd := range pendingDeletions {
			if pd.UserID == userID && pd.Endpoint == endpoint {
				if err := s.repo.RemovePendingDeletion(pd.UserID, pd.Endpoint, pd.Inbound); err != nil {
					s.logger.Warn("failed to remove pending deletion",
						slog.String("user_id", pd.UserID),
						slog.String("endpoint", pd.Endpoint),
						slog.String("inbound", pd.Inbound),
						slog.String("error", err.Error()),
					)
				} else {
					deleted++
				}
			}
		}
		
		if deleted == 0 {
			return fmt.Errorf("no pending deletion found for user_id=%s, endpoint=%s", userID, endpoint)
		}
		
		s.logger.Info("pending deletions removed",
			slog.String("user_id", userID),
			slog.String("endpoint", endpoint),
			slog.Int("count", deleted),
		)
		return nil
	}
	
	// Delete specific inbound
	err := s.repo.RemovePendingDeletion(userID, endpoint, inbound)
	if err != nil {
		return fmt.Errorf("failed to remove pending deletion: %w", err)
	}
	
	s.logger.Info("pending deletion removed",
		slog.String("user_id", userID),
		slog.String("endpoint", endpoint),
		slog.String("inbound", inbound),
	)
	
	return nil
}

// CleanupPendingDeletions attempts to delete users from nodes that previously failed
func (s *Service) CleanupPendingDeletions(ctx context.Context) (*models.CleanupResult, error) {
	pendingDeletions, err := s.repo.GetPendingDeletions()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending deletions: %w", err)
	}

	result := &models.CleanupResult{
		TotalAttempted: len(pendingDeletions),
		Errors:         []string{},
	}

	if len(pendingDeletions) == 0 {
		s.logger.Info("no pending deletions to clean up")
		return result, nil
	}

	s.logger.Info("starting cleanup of pending deletions", slog.Int("count", len(pendingDeletions)))

	for _, pd := range pendingDeletions {
		// Try to delete user from node
		err := s.nodeClient.DeleteUser(pd.Endpoint, pd.UserID)
		if err != nil {
			result.Failed++
			result.StillPending++
			errMsg := fmt.Sprintf("user=%s, endpoint=%s, inbound=%s: %s", pd.UserID, pd.Endpoint, pd.Inbound, err.Error())
			result.Errors = append(result.Errors, errMsg)
			
			s.logger.Warn("cleanup attempt failed",
				slog.String("user_id", pd.UserID),
				slog.String("endpoint", pd.Endpoint),
				slog.String("error", err.Error()),
			)
			
			// Update attempt counter
			if dbErr := s.repo.AddPendingDeletion(pd.UserID, pd.Endpoint, pd.Inbound, err.Error()); dbErr != nil {
				s.logger.Error("failed to update pending deletion", slog.String("error", dbErr.Error()))
			}
			continue
		}

		// Successful deletion - remove from pending and user_nodes
		result.Successful++
		
		if err := s.repo.RemovePendingDeletion(pd.UserID, pd.Endpoint, pd.Inbound); err != nil {
			s.logger.Warn("failed to remove pending deletion record", slog.String("error", err.Error()))
		}
		
		if err := s.repo.DeleteUserNode(pd.UserID, pd.Endpoint); err != nil {
			s.logger.Warn("failed to remove user_node record", slog.String("error", err.Error()))
		}
		
		s.logger.Info("cleanup successful",
			slog.String("user_id", pd.UserID),
			slog.String("endpoint", pd.Endpoint),
		)
	}

	s.logger.Info("cleanup completed",
		slog.Int("total", result.TotalAttempted),
		slog.Int("successful", result.Successful),
		slog.Int("failed", result.Failed),
		slog.Int("still_pending", result.StillPending),
	)

	return result, nil
}

