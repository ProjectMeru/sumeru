package auth

import go_qr "github.com/piglig/go-qr" // resolved to github.com/ProjectMeru/go-qr via go.mod replace

const totpQRQuietBorder = 4

// TOTPQRPng returns a PNG image for an otpauth URI.
func TOTPQRPng(otpauthURI string, size int) ([]byte, error) {
	if 0 >= size {
		size = 200
	}
	qr, err := go_qr.EncodeText(otpauthURI, go_qr.Medium)
	if err != nil {
		return nil, err
	}
	dim := qr.Size() + 2*totpQRQuietBorder
	scale := size / dim
	if scale < 1 {
		scale = 1
	}
	return qr.ToPNGBytes(go_qr.NewQrCodeImgConfig(scale, totpQRQuietBorder))
}
