package models

type Records struct {
	ID       int    `json:"id"`
	ItemID   string `json:"item_id"`
	Name     string `json:"name"`
	FolderID string `json:"folder_id"`
	Date     string `json:"date"`
}
