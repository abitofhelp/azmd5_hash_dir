package model

type LocalFile struct {
	pathInsideDirectory string
	base64Md5           string
}

func NewLocalFile(
	pathInsideDirectory string,
	base64Md5 string) *LocalFile {
	return &LocalFile{
		pathInsideDirectory: pathInsideDirectory,
		base64Md5:           base64Md5,
	}
}

func (x *LocalFile) PathInsideDirectory() string {
	return x.pathInsideDirectory
}

func (x *LocalFile) Base64Md5() string {
	return x.base64Md5
}
