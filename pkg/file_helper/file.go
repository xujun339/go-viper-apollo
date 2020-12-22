package file_helper

import (
	"os"
	"path"
	"strings"
)


// 判断文件是否存在
func Exists(fs os.File) (bool, error) {
	stat, err := fs.Stat()
	if err == nil {
		return !stat.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/**
判断路径 目录还是文件
目录 true
文件 false
 */
func IsDirOrFile(filePath string) (bool, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}


/**
device/sdk/CMakeLists.txt
return CMakeLists, .txt
 */
func FileBaseAndExt(filePath string) (filePrefix, fileSuffix string) {
	fileName := path.Base(filePath)
	ext := path.Ext(filePath)
	return strings.TrimSuffix(fileName, ext), ext
}



