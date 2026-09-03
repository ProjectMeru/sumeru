package modelmeta

func ExtractDefaultFromSelectionTailForTest(selection string) (trimmedSelection, defaultValue string) {
	return extractDefaultFromSelectionTail(selection)
}

func PeelSelectionTagForTest(tag string) (body, selection string) {
	return peelSelectionTag(tag)
}
