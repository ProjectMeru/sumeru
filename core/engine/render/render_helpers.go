package render

func splitFlashMessages(flashes []FlashMessage) (inline, toast []FlashMessage) {
	for _, f := range flashes {
		if f.ToastOnly {
			toast = append(toast, f)
			continue
		}
		inline = append(inline, f)
	}
	return inline, toast
}
