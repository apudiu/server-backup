package tasks

func ZipDirectory(sourceDir, destZipPath string) error {
	cmd := []string{
		"zip -r9",
		destZipPath,
		sourceDir,
	}
}
