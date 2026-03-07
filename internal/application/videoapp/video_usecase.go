package videoapp

type VideoUsecase struct {
	Upload  UploadVideoUsecase
	GetInfo GetVideoInfoUsecase
	Update  UpdateVideoUsecase
	Archive ArchiveVideoUsecase
	List    ListVideosUsecase
}
