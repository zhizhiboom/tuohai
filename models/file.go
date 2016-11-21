package models

type FileType int8

const (
	FileTypeImage      FileType = 1
	FileTypeDoc        FileType = 2
	FileTypeCompressed FileType = 3
	FileTypeSnippet    FileType = 4
)

type FileInfo struct {
	Id       string   `gorm:"column:id"`
	To       string   `gorm:"column:to"`
	Name     string   `gorm:"column:name"`
	Path     string   `gorm:"column:path"`
	Size     int      `gorm:"column:size"`
	Type     FileType `gorm:"column:type"`
	Ext      string   `gorm:"column:ext"`
	Category string   `gorm:"column:category"`
	Meta     *Image   `gorm:"column:meta"`
	Creator  string   `gorm:"column:creator"`
	Updated  int64    `gorm:"column:updated"`
	Created  int64    `gorm:"column:created"`
}

type Image struct {
}

func (file *FileInfo) TableName() string {
	return "file_info"
}

func (file *FileInfo) GetFileInfo() {
	if file.Type == FileTypeImage {
		file.GetImage()
	}
}

func (file *FileInfo) GetFilesInfo(to []string, ftype FileType) ([]FileInfo, error) {
	var infos []FileInfo
	err := db.Find(&infos, "to in (?) and type = ? and status = 0", to, ftype).Error
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(infos); i++ {
		if f.Type == FileTypeImage {
			infos[i].GetImage()
		}
	}
	return infos, err
}

func (file *FileInfo) GetImage() error {
	var img *Image
	err := db.Find(img, "id = ?", file.Id).Error
	if err != nil {
		return err
	}
	file.Meta = img
	return nil
}

func WriteFileToDB(file *FileInfo) {

}