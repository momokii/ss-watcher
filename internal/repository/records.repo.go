package repository

import (
	"database/sql"

	"github.com/momokii/ss-watcher/internal/models"
)

type RecordRepository interface {
	FindByName(tx *sql.Tx, filename string) (*models.Records, error)
	Create(tx *sql.Tx, record *models.Records) error
	Delete(tx *sql.Tx, id string) error
}

type recordRepository struct{}

func NewRecordsRepository() RecordRepository {
	return &recordRepository{}
}

func (r *recordRepository) FindByName(tx *sql.Tx, filename string) (*models.Records, error) {

	record := &models.Records{}

	err := tx.QueryRow("SELECT id, item_id, name, folder_id, date FROM records WHERE name = ?", filename).Scan(&record.ID, &record.ItemID, &record.Name, &record.FolderID, &record.Date)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (r *recordRepository) Create(tx *sql.Tx, record *models.Records) error {

	if _, err := tx.Exec("INSERT INTO records (item_id, name, folder_id, date) VALUES (?, ?, ?, ?)", record.ItemID, record.Name, record.FolderID, record.Date); err != nil {
		return err
	}

	return nil
}

func (r *recordRepository) Delete(tx *sql.Tx, id string) error {

	if _, err := tx.Exec("DELETE FROM records WHERE item_id = ?", id); err != nil {
		return err
	}

	return nil
}
