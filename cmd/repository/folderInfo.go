package repository

type FolderInfo struct {
	Inode    uint64 `json:"inode"`
	FullPath string `json:"full_path"`
	tags     []string
}
