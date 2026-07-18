package services

import (
	"errors"
	"testing"

	"LocalSpace/app/models"
)

func TestImportFileRequestKeywords(t *testing.T) {
	req := ImportFileRequest{
		FilePath:    "/test/path/file.txt",
		FileName:    "file.txt",
		Description: "test description",
		Tags:        []string{"tag1", "tag2"},
		Keywords:    "test keywords",
	}

	if req.Keywords != "test keywords" {
		t.Errorf("Expected Keywords to be 'test keywords', got '%s'", req.Keywords)
	}

	// Test default empty value
	emptyReq := ImportFileRequest{
		FilePath: "/test/path/file.txt",
		FileName: "file.txt",
	}

	if emptyReq.Keywords != "" {
		t.Errorf("Expected default Keywords to be empty string, got '%s'", emptyReq.Keywords)
	}
}

func TestNormalizeMetadataTags_TrimsDedupesAndDropsEmpty(t *testing.T) {
	service := &FileService{}

	got := service.normalizeMetadataTags([]string{" 工作 ", "", "重要", "工作", "  ", "归档 "})
	want := []string{"工作", "重要", "归档"}

	if len(got) != len(want) {
		t.Fatalf("expected %d tags, got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected tag %d to be %q, got %q", i, want[i], got[i])
		}
	}
}

func TestNormalizeMetadataDescription_TrimsWhitespace(t *testing.T) {
	service := &FileService{}

	got := service.normalizeMetadataDescription("  这是描述  ")
	if got != "这是描述" {
		t.Fatalf("expected trimmed description, got %q", got)
	}
}

type metadataRepoStub struct {
	findByIDCalledWith uint
	updatedID          uint
	updatedTags        []string
	updatedDescription string
	findErr            error
	updateErr          error
}

func (r *metadataRepoStub) FindByID(id uint) (*models.File, error) {
	r.findByIDCalledWith = id
	if r.findErr != nil {
		return nil, r.findErr
	}
	return &models.File{ID: id}, nil
}

func (r *metadataRepoStub) UpdateMetadata(id uint, tags []string, description string) error {
	r.updatedID = id
	r.updatedTags = append([]string{}, tags...)
	r.updatedDescription = description
	return r.updateErr
}

func TestUpdateFileMetadata_NormalizesInputBeforePersisting(t *testing.T) {
	repo := &metadataRepoStub{}
	service := &FileService{fileRepo: repo}

	err := service.UpdateFileMetadata(7, []string{" 工作 ", "工作", "重要", ""}, "  新描述  ")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if repo.findByIDCalledWith != 7 {
		t.Fatalf("expected FindByID to be called with 7, got %d", repo.findByIDCalledWith)
	}

	wantTags := []string{"工作", "重要"}
	if len(repo.updatedTags) != len(wantTags) {
		t.Fatalf("expected %d tags, got %d (%v)", len(wantTags), len(repo.updatedTags), repo.updatedTags)
	}
	for i := range wantTags {
		if repo.updatedTags[i] != wantTags[i] {
			t.Fatalf("expected tag %d to be %q, got %q", i, wantTags[i], repo.updatedTags[i])
		}
	}

	if repo.updatedDescription != "新描述" {
		t.Fatalf("expected trimmed description, got %q", repo.updatedDescription)
	}
}

func TestUpdateFileMetadata_PropagatesRepositoryErrors(t *testing.T) {
	repo := &metadataRepoStub{updateErr: errors.New("boom")}
	service := &FileService{fileRepo: repo}

	err := service.UpdateFileMetadata(9, []string{"tag"}, "desc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got, want := err.Error(), "failed to update file metadata: boom"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
