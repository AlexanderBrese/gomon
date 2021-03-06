package utils

import (
	"path/filepath"
)

const SOURCE_FILE = "test.go"
const SOURCE_FILE_CONTENT = `package main
	import "fmt"
	func main() {
		fmt.Println("hello world")
	} 
`

func PrepareBuild(srcDir string, buildDir string) error {
	if err := createSourceDir(srcDir); err != nil {
		return err
	}
	if err := createSourceFile(srcDir); err != nil {
		return err
	}
	return CreateBuildDir(buildDir)
}

func CleanupBuild(relSrcDir string, relBuildDir string) error {
	if err := removeSourceDir(relSrcDir); err != nil {
		return err
	}
	return RemoveBuildDir(relBuildDir)
}

func createSourceDir(srcDir string) error {
	return CreateDir(srcDir)
}

func createSourceFile(srcDir string) error {
	_, err := CreateFile(filepath.Join(srcDir, SOURCE_FILE), []byte(SOURCE_FILE_CONTENT))
	return err
}

func CreateBuildDir(buildDir string) error {
	return CreateDirIfNotExist(buildDir)
}

func RemoveBuildDir(relBuildDir string) error {
	return RemoveRootDir(relBuildDir)
}

func removeSourceDir(relSrcDir string) error {
	return RemoveRootDir(relSrcDir)
}
