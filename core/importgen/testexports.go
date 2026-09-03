package importgen

type ExportedRefForTest = exportedRef

type ScannedModelForTest = scannedModel

func RenderZRefsForTest(refs []ExportedRefForTest) string {
	return renderZRefs(refs)
}

func ParseDirModelsForTest(dir string) ([]ScannedModelForTest, error) {
	return parseDirModels(dir)
}

func ScanPackageForModelsForTest(dir string) ([]string, error) {
	return scanPackageForModels(dir)
}
