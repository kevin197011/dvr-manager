package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"dvr-manager/internal/repository"
	"dvr-manager/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理处理器（管理员）
type UserHandler struct {
	authService service.AuthService
	auditRepo   repository.AuditRepository
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(authService service.AuthService, auditRepo repository.AuditRepository) *UserHandler {
	return &UserHandler{authService: authService, auditRepo: auditRepo}
}

// CreateUserRequest 新增用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

// UpdateRoleRequest 修改角色请求
type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin user"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *UserHandler) audit(c *gin.Context, action, resource, detail, status string) {
	if h.auditRepo == nil {
		return
	}
	username, _ := c.Get("username")
	userStr, _ := username.(string)
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	_ = h.auditRepo.Insert(action, userStr, roleStr, c.ClientIP(), resource, detail, status)
}

// List 列出所有用户
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.authService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "list": users})
}

// Create 新增用户
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}
	u, err := h.authService.CreateUser(req.Username, req.Password, req.Role)
	if err != nil {
		h.audit(c, "user_create", req.Username, err.Error(), "fail")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	h.audit(c, "user_create", req.Username, fmt.Sprintf("创建用户（角色=%s）", req.Role), "success")
	c.JSON(http.StatusOK, gin.H{"success": true, "user": u})
}

// UpdateRole 修改用户角色
func (h *UserHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的用户 ID"})
		return
	}
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	// 防止管理员把自己降级（保证至少有一个 admin）
	target, err := h.authService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}
	currentUser, _ := c.Get("username")
	if currentUserStr, _ := currentUser.(string); currentUserStr == target.Username && req.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "不能修改自己的管理员角色"})
		return
	}

	if err := h.authService.UpdateUserRole(id, req.Role); err != nil {
		h.audit(c, "user_update_role", target.Username, err.Error(), "fail")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	h.audit(c, "user_update_role", target.Username, fmt.Sprintf("修改角色为 %s", req.Role), "success")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ResetPassword 重置某个用户的密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的用户 ID"})
		return
	}
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}
	target, err := h.authService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := h.authService.ResetPassword(id, req.NewPassword); err != nil {
		h.audit(c, "user_reset_password", target.Username, err.Error(), "fail")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	h.audit(c, "user_reset_password", target.Username, "重置密码", "success")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "密码已重置"})
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的用户 ID"})
		return
	}
	target, err := h.authService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 不允许删除自己
	currentUser, _ := c.Get("username")
	if currentUserStr, _ := currentUser.(string); currentUserStr == target.Username {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "不能删除当前登录的用户"})
		return
	}

	// 删除时确保至少保留一个管理员
	if target.Role == "admin" {
		users, err := h.authService.ListUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "无法验证管理员数量"})
			return
		}
		admins := 0
		for _, u := range users {
			if u.Role == "admin" {
				admins++
			}
		}
		if admins <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "至少保留一个管理员账号"})
			return
		}
	}

	if err := h.authService.DeleteUser(id); err != nil {
		h.audit(c, "user_delete", target.Username, err.Error(), "fail")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	h.audit(c, "user_delete", target.Username, "删除用户", "success")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
