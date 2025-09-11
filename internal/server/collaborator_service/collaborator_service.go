package collaborator_service

import (
	"clustta/internal/base_service"
	"clustta/internal/server/models"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func AddUserCollaborator(tx *sqlx.Tx, userId, collaboratorId string) error {
	id := uuid.New().String()
	params := map[string]interface{}{
		"id":              id,
		"user_id":         userId,
		"collaborator_id": collaboratorId,
	}
	err := base_service.Create(tx, "user_collaborator", params)
	if err != nil {
		return err
	}
	return nil
}

func GetUserCollaborations(tx *sqlx.Tx, userId string) ([]models.Studio, error) {
	query := `
		SELECT user.*
		FROM user
		LEFT JOIN
			user_collaborator ON user.id = user_collaborator.collaborator_id
		WHERE user_collaborator.collaborator_id = ?
	`
	userStudios := []models.Studio{}
	err := tx.Select(&userStudios, query, userId)
	if err != nil {
		return userStudios, err
	}
	return userStudios, nil
}

func GetUserCollaborators(tx *sqlx.Tx, userId string) ([]models.Studio, error) {
	query := `
		SELECT user.*
		FROM user
		LEFT JOIN
			user_collaborator ON user.id = user_collaborator.collaborator_id
		WHERE user_collaborator.user_id = ?
	`
	userStudios := []models.Studio{}
	err := tx.Select(&userStudios, query, userId)
	if err != nil {
		return userStudios, err
	}
	return userStudios, nil
}

func IsUserCollaborator(tx *sqlx.Tx, userId, collaboratorId string) (bool, error) {
	var isUserCollaborator bool
	query := `
		SELECT COUNT(*) > 0 AS user_collaborator
		FROM tomb
		WHERE user_id = ? AND collaborator_id = ?
	`
	err := tx.Get(&isUserCollaborator, query, userId, collaboratorId)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return isUserCollaborator, nil
}
