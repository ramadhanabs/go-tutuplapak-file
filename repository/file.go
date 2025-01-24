package repository

import (
	"database/sql"
	"fmt"

	"go-tutuplapak-file/db"
	"go-tutuplapak-file/model"
)

type FileRepository struct {
	DB *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{DB: db}
}

func (r *FileRepository) Create(originalFileUri string, compressedFileUri string) (model.File, error) {
	query := "INSERT INTO files (original_file_uri, compressed_file_uri) VALUES ($1, $2) RETURNING id, original_file_uri, compressed_file_uri"

	var user model.File
	err := db.DB.QueryRow(query, originalFileUri, compressedFileUri).Scan(&user.ID, &user.FileUri, &user.FileThumbnailUri)
	if err != nil {
		return model.File{}, fmt.Errorf("failed to create user: %v", err)
	}

	return user, nil
}
