package gitlab

type RepositoryFile struct {
	FileName        string `json:"file_name"`
	FilePath        string `json:"file_path"`
	Size            int    `json:"size"`
	Encoding        string `json:"encoding"`
	ContentSha256   string `json:"content_sha256"`
	Ref             string `json:"ref"`
	BlobID          string `json:"blob_id"`
	CommitID        string `json:"commit_id"`
	LastCommitID    string `json:"last_commit_id"`
	ExecuteFilemode bool   `json:"execute_filemode"`
	Content         string `json:"content"`
}
