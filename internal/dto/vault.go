package dto

import "time"

type (
	VaultFolderResponse struct {
		ID        string    `json:"id"`
		ParentID  *string   `json:"parentId"`
		Name      string    `json:"name"`
		Locked    bool      `json:"locked"`
		CreatedBy string    `json:"createdBy"`
		CreatedAt time.Time `json:"createdAt"`
	}

	VaultFileResponse struct {
		ID         string    `json:"id"`
		FolderID   *string   `json:"folderId"`
		Name       string    `json:"name"`
		Mime       string    `json:"mime"`
		Size       int64     `json:"size"`
		Locked     bool      `json:"locked"`
		UploadedBy string    `json:"uploadedBy"`
		CreatedAt  time.Time `json:"createdAt"`
	}

	VaultBreadcrumb struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	VaultBrowseResponse struct {
		Folder         *VaultFolderResponse  `json:"folder"`
		Breadcrumbs    []VaultBreadcrumb     `json:"breadcrumbs"`
		Folders        []VaultFolderResponse `json:"folders"`
		Files          []VaultFileResponse   `json:"files"`
		CanManageLocks bool                  `json:"canManageLocks"`
	}

	CreateVaultFolderRequest struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parentId"`
	}

	RenameVaultRequest struct {
		Name string `json:"name"`
	}
)
