package model

type LocalFile struct {
	pathInsideDirectory string
	azureMd5            string
}

func NewLocalFile(
	pathInsideDirectory string,
	azureMd5 string) *LocalFile {
	return &LocalFile{
		pathInsideDirectory: pathInsideDirectory,
		azureMd5:            azureMd5,
	}
}

func (x *LocalFile) PathInsideDirectory() string {
	return x.pathInsideDirectory
}

func (x *LocalFile) AzureMd5() string {
	return x.azureMd5
}
