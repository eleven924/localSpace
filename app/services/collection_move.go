package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
)

func moveFileToCollection(
	fileRepo *repositories.FileRepository,
	storageService *StorageService,
	collectionRepo collectionResolver,
	file *models.File,
	collectionID *uint,
) error {
	targetName := ""
	if collectionID != nil {
		collection, err := collectionRepo.FindByID(*collectionID)
		if err != nil {
			return fmt.Errorf("failed to resolve collection: %w", err)
		}
		targetName = collection.Name
	}

	oldPath := file.FilePath
	master, err := findMasterForPath(storageService, oldPath)
	if err != nil {
		return err
	}

	layout, err := storageService.GetStorageLayoutConfig()
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

	_, subPath, err := storageService.EnsureSubDirectory(master.ID, file.FileType)
	if err != nil {
		return err
	}
	if layout.Strategy == "type_collection" {
		subPath = filepath.Join(subPath, segment)
	}
	newPath := filepath.Join(subPath, file.FileName)

	if newPath == oldPath {
		file.CollectionID = collectionID
		return fileRepo.Update(file)
	}

	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("destination already exists: %s", newPath)
	}
	if err := storageService.EnsureStorageDirExists(filepath.Dir(newPath)); err != nil {
		return err
	}

	file.CollectionID = collectionID
	file.FilePath = newPath
	if err := fileRepo.Update(file); err != nil {
		return err
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		file.CollectionID = nil
		file.FilePath = oldPath
		_ = fileRepo.Update(file)
		return fmt.Errorf("failed to move file: %w", err)
	}

	if err := storageService.UpdateMasterDirectorySize(master.ID, file.FileType, -file.FileSize); err != nil {
		fmt.Printf("Warning: failed to update old master size: %v\n", err)
	}
	newMaster, err := findMasterForPath(storageService, newPath)
	if err == nil && newMaster.ID != master.ID {
		if err := storageService.UpdateMasterDirectorySize(newMaster.ID, file.FileType, file.FileSize); err != nil {
			fmt.Printf("Warning: failed to update new master size: %v\n", err)
		}
	}
	return nil
}

func findMasterForPath(storageService *StorageService, path string) (*models.StorageDir, error) {
	masters, err := storageService.GetMasterDirectories()
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
