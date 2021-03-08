package utils

// CreateBuildDirIfNotExist creates the build directory if it does not already exist
func CreateBuildDirIfNotExist(buildDir string) error {
	return CreateAllDirIfNotExist(buildDir)
}

// RemoveRootBuildDir removes the given build dir at it's root
func RemoveRootBuildDir(relBuildDir string) error {
	return RemoveRootDir(relBuildDir)
}
