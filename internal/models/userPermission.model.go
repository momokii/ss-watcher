package models

type UserPermission struct {
	ID           int    `json:"id"`
	PermissionID string `json:"permission_id"`
	Email        string `json:"email"`
}
