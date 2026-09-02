package report

// ExportTemplatePDFWithLayout builds a branded document PDF using DocumentLayout.
func ExportTemplatePDFWithLayout(in TemplatePDFInput, layout DocumentLayout) ([]byte, error) {
	if layout.PaperFormat != "" {
		in.PageSize = layout.PaperFormat
	}
	data, err := ExportTemplatePDF(in)
	if err != nil {
		return nil, err
	}
	_ = layout
	return data, nil
}
