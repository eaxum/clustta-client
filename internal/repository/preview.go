package repository

import (
	"bytes"
	"clustta/internal/base_service"
	"clustta/internal/constants"
	"clustta/internal/error_service"
	"clustta/internal/repository/models"
	"clustta/internal/repository/repositorypb"
	"clustta/internal/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/DataDog/zstd"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

func CreatePreview(tx *sqlx.Tx, previewPath string) (models.Preview, error) {
	previewFileExtension := ""
	preview := models.Preview{}

	if previewPath == "" {
		return preview, errors.New("preview file does not exist")
	}
	if previewPath != "" {
		if _, err := os.Stat(previewPath); os.IsNotExist(err) {
			return preview, errors.New("preview file does not exist")
		}
	}

	hash, err := utils.GenerateXXHashChecksum(previewPath)
	if err != nil {
		return preview, err
	}
	preview, err = GetPreview(tx, hash)
	if err == nil {
		return preview, nil
	}
	previewFileExtension = filepath.Ext(previewPath)
	fileData, err := os.ReadFile(previewPath)
	if err != nil {
		return preview, err
	}
	_, err = tx.Exec("INSERT INTO preview (hash, preview, extension) VALUES (?, ?, ?)",
		hash,
		fileData,
		previewFileExtension,
	)
	if err != nil {
		return preview, err
	}

	preview, err = GetPreview(tx, hash)
	if err != nil {
		return preview, err
	}
	return preview, nil
}
func AddPreview(tx *sqlx.Tx, hash string, preview []byte, extension string) error {
	_, err := GetPreview(tx, hash)
	if err == nil {
		return nil
	}
	_, err = tx.Exec("INSERT INTO preview (hash, preview, extension) VALUES (?, ?, ?)",
		hash,
		preview,
		extension,
	)
	if err != nil {
		return err
	}
	return nil
}

func AddPreviews(tx *sqlx.Tx, previews []models.Preview) error {
	for _, preview := range previews {
		_, err := GetPreview(tx, preview.Hash)
		if err == nil {
			return nil
		} else {
			if err == error_service.ErrPreviewNotFound {
				_, err = tx.Exec("INSERT INTO preview (hash, preview, extension) VALUES (?, ?, ?)",
					preview.Hash,
					preview.Preview,
					preview.Extension,
				)
				if err != nil {
					return err
				}
			}
			return err
		}

	}
	return nil
}

func GetPreview(tx *sqlx.Tx, hash string) (models.Preview, error) {
	preview := models.Preview{}
	query := "SELECT * FROM 'preview' WHERE hash = ?"
	err := tx.Get(&preview, query, hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return preview, error_service.ErrPreviewNotFound
		}
		return preview, err
	}
	return preview, nil
}

func AddEntityPreview(tx *sqlx.Tx, entityId, entityModel, previewPath string) (models.Preview, error) {
	preview, err := CreatePreview(tx, previewPath)
	if err != nil {
		return preview, err
	}

	params := map[string]any{
		"preview_id": preview.Hash,
	}
	err = base_service.Update(tx, entityModel, entityId, params)
	if err != nil {
		return preview, err
	}
	err = base_service.UpdateMtime(tx, entityModel, entityId, utils.GetEpochTime())
	if err != nil {
		return preview, err
	}
	return preview, nil
}
func SetEntityPreview(tx *sqlx.Tx, entityId, entityModel, previewHash string) error {
	params := map[string]any{
		"preview_id": previewHash,
	}
	err := base_service.Update(tx, entityModel, entityId, params)
	if err != nil {
		return err
	}
	err = base_service.UpdateMtime(tx, entityModel, entityId, utils.GetEpochTime())
	if err != nil {
		return err
	}
	return nil
}

func PreviewExists(hash string, tx *sqlx.Tx) bool {
	var previewId string
	tx.Get(&previewId, "SELECT hash FROM preview WHERE hash = ?", hash)
	return previewId != ""
}

func GetNonExistingPreviews(tx *sqlx.Tx, previewIds []string) ([]string, error) {
	var nonExistentPreviews []string
	for _, previewId := range previewIds {
		if PreviewExists(previewId, tx) {
			continue
		}
		nonExistentPreviews = append(nonExistentPreviews, previewId)
	}

	return nonExistentPreviews, nil
}

func PullPreviews(tx *sqlx.Tx, remoteUrl string, previewHashes []string, callback func(int, int, string, string)) error {
	dataUrl := remoteUrl + "/previews"
	client := &http.Client{}
	totalPreviews := len(previewHashes)
	processedPreviews := 0
	if utils.IsValidURL(remoteUrl) {
		for _, previewHash := range previewHashes {
			data := map[string]any{
				"previews": []string{previewHash},
			}
			jsonData, err := json.Marshal(data)
			if err != nil {
				return err
			}

			req, err := http.NewRequest("GET", dataUrl, bytes.NewBuffer(jsonData))
			if err != nil {
				return err
			}
			req.Header.Set("Clustta-Agent", constants.USER_AGENT)
			response, err := client.Do(req)
			if err != nil {
				return err
			}
			defer response.Body.Close()
			responseCode := response.StatusCode
			if responseCode == 200 {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					return fmt.Errorf("error reading response body: %s", err.Error())
				}

				decompressedData, err := zstd.Decompress(nil, body)
				if err != nil {
					return err
				}

				previewList := repositorypb.Previews{}
				err = proto.Unmarshal(decompressedData, &previewList)
				if err != nil {
					return err
				}
				previews := FromPbPreviews(previewList.Previews)

				err = AddPreviews(tx, previews)
				if err != nil {
					return fmt.Errorf("error writing preview: %s", err.Error())
				}
				processedPreviews++
				message := fmt.Sprintf("Pulling Preview %d/%d", processedPreviews, totalPreviews)
				callback(processedPreviews, totalPreviews, message, "")
			} else if responseCode == 400 {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					return err
				}
				return errors.New(string(body))
			} else {

				return errors.New("unknown error while fetching data")
			}
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return err
		}
		defer dbConn.Close()
		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return err
		}
		defer remoteTx.Rollback()
		for _, previewHash := range previewHashes {
			var preview models.Preview
			err = remoteTx.Get(&preview, "SELECT * FROM preview WHERE hash = ?", previewHash)
			if err != nil {
				return err
			}
			_, err = tx.Exec("INSERT INTO preview (hash, preview, extension) VALUES (?, ?, ?)",
				preview.Hash,
				preview.Preview,
				preview.Extension,
			)
			if err != nil {
				return err
			}
			processedPreviews++
			message := fmt.Sprintf("Pulling Preview %d/%d", processedPreviews, totalPreviews)
			callback(processedPreviews, totalPreviews, message, "")
		}
	}
	return nil
}

func PushPreviews(tx *sqlx.Tx, remoteUrl string, userId string, previewHashes []string, callback func(int, int, string, string)) error {
	dataUrl := remoteUrl + "/previews"
	client := &http.Client{}
	totalPreviews := len(previewHashes)
	processedPreviews := 0

	if utils.IsValidURL(remoteUrl) {
		for _, previewHash := range previewHashes {
			previews := []models.Preview{}
			var preview models.Preview
			err := tx.Get(&preview, "SELECT * FROM preview WHERE hash = ?", previewHash)
			if err != nil {
				return err
			}
			previews = append(previews, preview)

			pbPreviews := ToPbPreviews(previews)
			pbPreviewsList := &repositorypb.Previews{Previews: pbPreviews}

			pbPreviewsListByte, err := proto.Marshal(pbPreviewsList)
			if err != nil {
				return err
			}

			compressedData, err := zstd.CompressLevel(nil, pbPreviewsListByte, 3)
			if err != nil {
				return err
			}

			req, err := http.NewRequest("POST", dataUrl, bytes.NewBuffer(compressedData))
			if err != nil {
				return err
			}
			req.Header.Set("Clustta-Agent", constants.USER_AGENT)

			response, err := client.Do(req)
			if err != nil {
				return err
			}
			defer response.Body.Close()
			responseCode := response.StatusCode
			if responseCode == 200 {
				processedPreviews++
				message := fmt.Sprintf("Pushing Preview %d/%d", processedPreviews, totalPreviews)
				callback(processedPreviews, totalPreviews, message, "")
			} else if responseCode == 400 {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					return err
				}
				return errors.New(string(body))
			} else {
				return errors.New("unknown error while pushing previews")
			}
		}
	} else if utils.FileExists(remoteUrl) {
		dbConn, err := utils.OpenDb(remoteUrl)
		if err != nil {
			return err
		}
		defer dbConn.Close()
		remoteTx, err := dbConn.Beginx()
		if err != nil {
			return err
		}

		for _, previewHash := range previewHashes {
			var preview models.Preview
			err = tx.Get(&preview, "SELECT * FROM preview WHERE hash = ?", previewHash)
			if err != nil {
				return err
			}
			_, err = remoteTx.Exec("INSERT INTO preview (hash, preview, extension) VALUES (?, ?, ?)",
				preview.Hash,
				preview.Preview,
				preview.Extension,
			)
			if err != nil {
				return err
			}
			processedPreviews++
			message := fmt.Sprintf("Pushing Preview %d/%d", processedPreviews, totalPreviews)
			callback(processedPreviews, totalPreviews, message, "")
		}
		err = remoteTx.Commit()
		if err != nil {
			remoteTx.Rollback()
			return err
		}
	}

	return nil
}
