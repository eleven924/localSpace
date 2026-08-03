package services

import (
	"fmt"
	"os"

	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
)

// CollectionService manages collections and batch file operations.
type CollectionService struct {
	collectionRepo *repositories.CollectionRepository
	fileRepo       *repositories.FileRepository
	storageService *StorageService
}

// NewCollectionService creates a new CollectionService.
func NewCollectionService(
	collectionRepo *repositories.CollectionRepository,
	fileRepo *repositories.FileRepository,
	storageService *StorageService,
) *CollectionService {
	return &CollectionService{
		collectionRepo: collectionRepo,
		fileRepo:       fileRepo,
		storageService: storageService,
	}
}

// GetCollections returns all collections ordered by name.
func (s *CollectionService) GetCollections() ([]models.Collection, error) {
	return s.collectionRepo.GetAll()
}

func (s *CollectionService) GetCollectionFilterCounts(fileType string) (*models.CollectionFilterCounts, error) {
	return s.fileRepo.CountCollectionFilters(fileType)
}

// AddCollection creates a new collection.
func (s *CollectionService) AddCollection(name string) (uint, error) {
	return s.collectionRepo.Add(name)
}

// RemoveCollection removes a collection if it is not referenced by any file.
func (s *CollectionService) RemoveCollection(id uint) error {
	count, err := s.collectionRepo.CountFilesByCollectionID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("collection is referenced by %d file(s)", count)
	}
	return s.collectionRepo.Remove(id)
}

// BatchUpdateFilesCollection updates the collection for multiple files and moves them on disk if needed.
func (s *CollectionService) BatchUpdateFilesCollection(ids []uint, collectionID *uint) (models.BatchMoveResult, error) {
	var result models.BatchMoveResult
	for _, id := range ids {
		file, err := s.fileRepo.FindByID(id)
		if err != nil {
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchMoveFailedItem{FileID: id, Error: err.Error()})
			continue
		}

		if err := moveFileToCollection(s.fileRepo, s.storageService, s.collectionRepo, file, collectionID); err != nil {
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchMoveFailedItem{FileID: id, FileName: file.FileName, Error: err.Error()})
			continue
		}
		result.SuccessCount++
	}
	return result, nil
}

// BatchDeleteFiles deletes multiple files and their on-disk records.
func (s *CollectionService) BatchDeleteFiles(ids []uint) (models.BatchDeleteResult, error) {
	var result models.BatchDeleteResult
	for _, id := range ids {
		file, err := s.fileRepo.FindByID(id)
		if err != nil {
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchDeleteFailedItem{FileID: id, Error: err.Error()})
			continue
		}
		if err := s.fileRepo.Delete(id); err != nil {
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchDeleteFailedItem{FileID: id, FileName: file.FileName, Error: err.Error()})
			continue
		}
		if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchDeleteFailedItem{FileID: id, FileName: file.FileName, Error: err.Error()})
			continue
		}
		if err := s.storageService.UpdateStorageSize(file.FileType, -file.FileSize); err != nil {
			fmt.Printf("Warning: failed to update storage size: %v\n", err)
		}
		result.SuccessCount++
	}
	return result, nil
}
