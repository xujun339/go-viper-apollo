package viper_helper

type SourceType int32
const(
	LOCAL_FILE SourceType = iota // 本地文件
	REMOTE_APOLLO // 远端apollo
)

func (that SourceType) GetDescript() string {
	switch that {
		case LOCAL_FILE : return "文件"
		case REMOTE_APOLLO : return "阿波罗"
	}
	return ""
}

type FileType int32
const(
	YML FileType = iota
	JSON
	PROPERTIES
	NODEFINE
)

var FileTypeMap map[FileType]string

func init() {
	FileTypeMap = toMap()
}

func toMap() map[FileType]string {
	fileTypeMap := map[FileType]string{
		YML : "yml",
		JSON : "json",
		PROPERTIES : "properties",
	}
	return fileTypeMap
}

func (that FileType) String() string {
	if v, ok := FileTypeMap[that]; ok {
		return v
	}
	return ""
}


func ToFileType(fileType string) FileType {
	for key, value := range FileTypeMap {
		if value == fileType {
			return key
		}
	}
	return NODEFINE
}
