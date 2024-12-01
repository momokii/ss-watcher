package repository

import (
	"database/sql"
	"strings"

	"github.com/momokii/ss-watcher/internal/models"
)

type UserPermission interface {
	FindByID(tx *sql.Tx, permission_ids []string) (*[]models.UserPermission, error)
	Create(tx *sql.Tx, permission *models.UserPermission) error
	Delete(tx *sql.Tx, permission_id string) error
}

type userPermission struct{}

func NewUserPermission() UserPermission {
	return &userPermission{}
}

func (r *userPermission) FindByID(tx *sql.Tx, permission_ids []string) (*[]models.UserPermission, error) {

	var permissions []models.UserPermission

	query := "SELECT id, permission_id, email FROM user_permission WHERE permission_id IN (" + strings.Join(permission_ids, ",") + ")"

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var permission models.UserPermission

		if err := rows.Scan(&permission.ID, &permission.PermissionID, &permission.Email); err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	return &permissions, nil
}

func (r *userPermission) Create(tx *sql.Tx, permission *models.UserPermission) error {

	if _, err := tx.Exec("INSERT INTO user_permission (permission_id, email) VALUES (?, ?)", permission.PermissionID, permission.Email); err != nil {
		return err
	}

	return nil
}

func (r *userPermission) Delete(tx *sql.Tx, permission_id string) error {

	if _, err := tx.Exec("DELETE FROM user_permission WHERE permission_id = ?", permission_id); err != nil {
		return err
	}

	return nil
}
