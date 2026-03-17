package usecase

import (
	"context"
	"fmt"
	"time"

	"magua-fly-golang/internal/repository"
)

// ContractFilesUseCase 封装合同文件上传/下载的核心业务逻辑。
type ContractFilesUseCase struct {
	contractRepo     repository.ContractRepository
	contractFileRepo repository.ContractFileMetaRepository
	gridFS           GridFSBucket
}

// GridFSBucket 是对 Mongo GridFS 的抽象，便于测试与替换实现。
type GridFSBucket interface {
	UploadFromStream(ctx context.Context, filename string, data []byte, metadata map[string]interface{}) (string, error)
	DownloadToStream(ctx context.Context, fileID string) ([]byte, map[string]interface{}, error)
}

// NewContractFilesUseCase 创建合同文件用例。
func NewContractFilesUseCase(
	contractRepo repository.ContractRepository,
	contractFileRepo repository.ContractFileMetaRepository,
	gridFS GridFSBucket,
) *ContractFilesUseCase {
	return &ContractFilesUseCase{
		contractRepo:     contractRepo,
		contractFileRepo: contractFileRepo,
		gridFS:           gridFS,
	}
}

// UploadContractFileInput 上传合同文件入参。
type UploadContractFileInput struct {
	ContractID string
	StudentID  string
	UploaderID string
	Filename   string
	ContentType string
	Data       []byte
	Tags       []string
}

// ContractFileDTO 返回给前端的合同文件信息。
type ContractFileDTO struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	ContentType string   `json:"content_type"`
	Size       int64     `json:"size"`
	Tags       []string  `json:"tags,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// UploadContractFile 上传合同扫描件：写入 GridFS、记录 Mongo 元数据，并更新 PG contracts.mongo_file_ids。
func (u *ContractFilesUseCase) UploadContractFile(ctx context.Context, in *UploadContractFileInput) (*ContractFileDTO, error) {
	if in == nil {
		return nil, fmt.Errorf("input is nil")
	}
	// 校验合同存在
	contract, err := u.contractRepo.GetByID(ctx, in.ContractID)
	if err != nil {
		return nil, err
	}
	if contract.StudentID != in.StudentID {
		return nil, fmt.Errorf("contract student mismatch")
	}

	// 1) 写入 GridFS
	gridMeta := map[string]interface{}{
		"contract_id": in.ContractID,
		"student_id":  in.StudentID,
		"uploader_id": in.UploaderID,
		"content_type": in.ContentType,
	}
	gridID, err := u.gridFS.UploadFromStream(ctx, in.Filename, in.Data, gridMeta)
	if err != nil {
		return nil, err
	}

	// 2) 写入 Mongo 元数据集合
	now := time.Now()
	meta := &repository.ContractFileMeta{
		StudentID:   in.StudentID,
		ContractID:  in.ContractID,
		Filename:    in.Filename,
		ContentType: in.ContentType,
		Size:        int64(len(in.Data)),
		GridFSID:    gridID,
		Tags:        in.Tags,
		UploadedAt:  now.Unix(),
		UploadedBy:  in.UploaderID,
	}
	if err := u.contractFileRepo.Insert(ctx, meta); err != nil {
		return nil, err
	}

	// 3) 更新 PG 合同记录中的 mongo_file_ids（追加）
	contract.MongoFileIDs = append(contract.MongoFileIDs, meta.ID)
	if err := u.contractRepo.Update(ctx, contract); err != nil {
		return nil, err
	}

	return &ContractFileDTO{
		ID:          meta.ID,
		Filename:    meta.Filename,
		ContentType: meta.ContentType,
		Size:        meta.Size,
		Tags:        meta.Tags,
		UploadedAt:  now,
	}, nil
}

// ListContractFiles 返回某合同下的所有文件（仅元数据）。
func (u *ContractFilesUseCase) ListContractFiles(ctx context.Context, contractID string) ([]*ContractFileDTO, error) {
	metas, err := u.contractFileRepo.ListByContract(ctx, contractID)
	if err != nil {
		return nil, err
	}
	var out []*ContractFileDTO
	for _, m := range metas {
		out = append(out, &ContractFileDTO{
			ID:          m.ID,
			Filename:    m.Filename,
			ContentType: m.ContentType,
			Size:        m.Size,
			Tags:        m.Tags,
			UploadedAt:  time.Unix(m.UploadedAt, 0),
		})
	}
	return out, nil
}

// DownloadContractFile 下载合同文件，返回二进制与内容类型。
func (u *ContractFilesUseCase) DownloadContractFile(ctx context.Context, fileMetaID string) ([]byte, string, error) {
	if fileMetaID == "" {
		return nil, "", fmt.Errorf("file id required")
	}
	meta, err := u.contractFileRepo.GetByID(ctx, fileMetaID)
	if err != nil {
		return nil, "", err
	}
	data, _, err := u.gridFS.DownloadToStream(ctx, meta.GridFSID)
	if err != nil {
		return nil, "", err
	}
	return data, meta.ContentType, nil
}

