package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"vpnbot-core-go/internal/models"
	"vpnbot-core-go/internal/nodeclient"
	"vpnbot-core-go/internal/repository"
	"vpnbot-core-go/internal/service"
)

type Handlers struct {
	repo       *repository.Repository
	service    *service.Service
	nodeClient *nodeclient.Client
	logger     *slog.Logger
}

func NewHandlers(repo *repository.Repository, svc *service.Service, nodeClient *nodeclient.Client, logger *slog.Logger) *Handlers {
	return &Handlers{
		repo:       repo,
		service:    svc,
		nodeClient: nodeClient,
		logger:     logger,
	}
}

// Health check
func (h *Handlers) Health(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

// CreateUser creates a user on all active nodes or specific node
func (h *Handlers) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	// Body is optional - if empty, creates on all nodes
	_ = c.BodyParser(&req)
	
	result, err := h.service.CreateUser(c.Context(), req.UUID, req.Endpoint)
	if err != nil {
		h.logger.Error("create user failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// DeleteUser deletes a user from all nodes or specific node (backward compatible)
func (h *Handlers) DeleteUser(c *fiber.Ctx) error {
	var req models.DeleteUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.ID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing user ID",
		})
	}

	// Smart detection: backward compatibility
	if req.Endpoint != "" {
		// Old API: delete from specific node only
		h.logger.Info("deleting user from specific node (legacy mode)",
			slog.String("user_id", req.ID),
			slog.String("endpoint", req.Endpoint),
		)
		if err := h.service.DeleteUserFromNode(c.Context(), req.ID, req.Endpoint); err != nil {
			h.logger.Error("delete user from node failed", slog.String("error", err.Error()))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	} else {
		// New API: delete from all tracked nodes
		if err := h.service.DeleteUser(c.Context(), req.ID); err != nil {
			h.logger.Error("delete user failed", slog.String("error", err.Error()))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

// GetServers returns all active servers
func (h *Handlers) GetServers(c *fiber.Ctx) error {
	servers, err := h.repo.GetServers(true)
	if err != nil {
		h.logger.Error("get servers failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(servers)
}

// PostServers creates or replaces servers
func (h *Handlers) PostServers(c *fiber.Ctx) error {
	var req models.ServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Servers) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing servers in request",
		})
	}

	// Set default inbound type if not specified
	for i := range req.Servers {
		if req.Servers[i].InboundType == "" {
			req.Servers[i].InboundType = "trojan"
		}
	}

	if err := h.repo.CreateServers(req.Servers); err != nil {
		h.logger.Error("create servers failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusCreated)
}

// PutServers updates existing servers
func (h *Handlers) PutServers(c *fiber.Ctx) error {
	var req models.ServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Servers) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing servers in request",
		})
	}

	if err := h.repo.UpdateServers(req.Servers); err != nil {
		h.logger.Error("update servers failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteServers deletes servers
func (h *Handlers) DeleteServers(c *fiber.Ctx) error {
	var req models.ServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Servers) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing servers in request",
		})
	}

	endpoints := make([]string, len(req.Servers))
	for i, s := range req.Servers {
		endpoints[i] = s.Endpoint
	}

	if err := h.repo.DeleteServers(endpoints); err != nil {
		h.logger.Error("delete servers failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

// GetUsers gets users from a specific node
func (h *Handlers) GetUsers(c *fiber.Ctx) error {
	var req models.NodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Endpoint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing endpoint",
		})
	}

	result, err := h.nodeClient.GetUsers(req.Endpoint)
	if err != nil {
		h.logger.Error("get users failed",
			slog.String("endpoint", req.Endpoint),
			slog.String("error", err.Error()),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result.Users)
}

// GetInbounds gets inbound configurations from a specific node
func (h *Handlers) GetInbounds(c *fiber.Ctx) error {
	var req models.NodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Endpoint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing endpoint",
		})
	}

	result, err := h.nodeClient.GetInbounds(req.Endpoint)
	if err != nil {
		h.logger.Error("get inbounds failed",
			slog.String("endpoint", req.Endpoint),
			slog.String("error", err.Error()),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// GetStats returns comprehensive statistics for all nodes
func (h *Handlers) GetStats(c *fiber.Ctx) error {
	stats, err := h.service.GetStats(c.Context())
	if err != nil {
		h.logger.Error("get stats failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stats)
}

// GetUsersCount returns total users count
func (h *Handlers) GetUsersCount(c *fiber.Ctx) error {
	count, err := h.service.GetUsersCount(c.Context())
	if err != nil {
		h.logger.Error("get users count failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(models.UsersCountResponse{
		Count: count,
	})
}

// GetAllUsers returns all users with their nodes information
func (h *Handlers) GetAllUsers(c *fiber.Ctx) error {
	users, err := h.service.GetAllUsers(c.Context())
	if err != nil {
		h.logger.Error("get all users failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(users)
}

// GetNodesWithUsers returns all nodes with their users information
func (h *Handlers) GetNodesWithUsers(c *fiber.Ctx) error {
	nodes, err := h.service.GetNodesWithUsers(c.Context())
	if err != nil {
		h.logger.Error("get nodes with users failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(nodes)
}

// GetPendingDeletionsHandler handles GET /api/cleanup/pending
func (h *Handlers) GetPendingDeletionsHandler(c *fiber.Ctx) error {
	pendingDeletions, err := h.service.GetPendingDeletions()
	if err != nil {
		h.logger.Error("failed to get pending deletions", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(models.PendingDeletionsResponse{
		Count:            len(pendingDeletions),
		PendingDeletions: pendingDeletions,
	})
}

// CleanupPendingDeletionsHandler handles POST /api/cleanup
func (h *Handlers) CleanupPendingDeletionsHandler(c *fiber.Ctx) error {
	result, err := h.service.CleanupPendingDeletions(c.Context())
	if err != nil {
		h.logger.Error("cleanup failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// DeletePendingDeletionHandler handles DELETE /api/cleanup/pending
func (h *Handlers) DeletePendingDeletionHandler(c *fiber.Ctx) error {
	var req models.DeletePendingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.UserID == "" || req.Endpoint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId and endpoint are required",
		})
	}

	if err := h.service.DeletePendingDeletion(req.UserID, req.Endpoint, req.Inbound); err != nil {
		h.logger.Error("failed to delete pending deletion", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

