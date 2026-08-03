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
		targetName = strings.TrimSpace(collection.Name)
	}

	oldPath := file.FilePath
	oldCollectionID := file.CollectionID
	oldCollectionName := file.CollectionName
	master, err := findMasterForPath(storageService, oldPath)
	if err != nil {
		return err
	}

	// 目标路径复用导入时的存储布局规则，保证移动合集和新导入落在同一类目录结构下。
	newPath, err := storageService.GetStoragePathForFileWithMaster(master.ID, file.FileType, targetName, file.FileName)
	if err != nil {
		return err
	}

	if newPath == oldPath {
		file.CollectionID = collectionID
		file.CollectionName = targetName
		return fileRepo.UpdateCollectionMove(file.ID, collectionID, targetName, oldPath)
	}

	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("destination already exists: %s", newPath)
	}
	if err := storageService.EnsureStorageDirExists(filepath.Dir(newPath)); err != nil {
		return err
	}

	file.CollectionID = collectionID
	file.CollectionName = targetName
	file.FilePath = newPath
	if err := fileRepo.UpdateCollectionMove(file.ID, collectionID, targetName, newPath); err != nil {
		return err
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		file.CollectionID = oldCollectionID
		file.CollectionName = oldCollectionName
		file.FilePath = oldPath
		_ = fileRepo.UpdateCollectionMove(file.ID, oldCollectionID, oldCollectionName, oldPath)
		return fmt.Errorf("failed to move file: %w", err)
	}

	newMaster, err := findMasterForPath(storageService, newPath)
	if err == nil && newMaster.ID != master.ID {
		if err := storageService.UpdateMasterDirectorySize(master.ID, file.FileType, -file.FileSize); err != nil {
			fmt.Printf("Warning: failed to update old master size: %v\n", err)
		}
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
