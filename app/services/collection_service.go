package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	targetName := ""
	if collectionID != nil {
		c, err := s.collectionRepo.FindByID(*collectionID)
		if err != nil {
			return result, err
		}
		targetName = c.Name
	}

	for _, id := range ids {
		file, err := s.fileRepo.FindByID(id)
		if err != nil {
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchMoveFailedItem{FileID: id, Error: err.Error()})
			continue
		}

		if err := s.moveFileToCollection(file, collectionID, targetName); err != nil {
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchMoveFailedItem{FileID: id, FileName: file.FileName, Error: err.Error()})
			continue
		}
		result.SuccessCount++
	}
	return result, nil
}

func (s *CollectionService) moveFileToCollection(file *models.File, collectionID *uint, targetName string) error {
	oldPath := file.FilePath
	master, err := s.findMasterForPath(oldPath)
	if err != nil {
		return err
	}

	layout, err := s.storageService.GetStorageLayoutConfig()
	if err != nil {
		return err
	}

	segment := targetName
	if segment == "" {
		segment = layout.UnsortedFolderName
	}
	if layout.SanitizeFolderName {
		segment = sanitizePathSegment(segment)
	}

	_, subPath, err := s.storageService.EnsureSubDirectory(master.ID, file.FileType)
	if err != nil {
		return err
	}
	if layout.Strategy == "type_collection" {
		subPath = filepath.Join(subPath, segment)
	}
	newPath := filepath.Join(subPath, file.FileName)

	if newPath == oldPath {
		file.CollectionID = collectionID
		return s.fileRepo.Update(file)
	}

	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("destination already exists: %s", newPath)
	}
	if err := s.storageService.EnsureStorageDirExists(filepath.Dir(newPath)); err != nil {
		return err
	}

	file.CollectionID = collectionID
	file.FilePath = newPath
	if err := s.fileRepo.Update(file); err != nil {
		return err
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		file.CollectionID = nil
		file.FilePath = oldPath
		_ = s.fileRepo.Update(file)
		return fmt.Errorf("failed to move file: %w", err)
	}

	if err := s.storageService.UpdateMasterDirectorySize(master.ID, file.FileType, -file.FileSize); err != nil {
		fmt.Printf("Warning: failed to update old master size: %v\n", err)
	}
	newMaster, err := s.findMasterForPath(newPath)
	if err == nil && newMaster.ID != master.ID {
		if err := s.storageService.UpdateMasterDirectorySize(newMaster.ID, file.FileType, file.FileSize); err != nil {
			fmt.Printf("Warning: failed to update new master size: %v\n", err)
		}
	}
	return nil
}

func (s *CollectionService) findMasterForPath(path string) (*models.StorageDir, error) {
	masters, err := s.storageService.GetMasterDirectories()
	if err != nil {
		return nil, err
	}
	for i := range masters {
		if strings.HasPrefix(path, masters[i].Path) {
			return &masters[i], nil
		}
	}
	return nil, fmt.Errorf("no master directory contains path %s", path)
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
