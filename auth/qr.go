package auth

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"github.com/bogdanrat/web-server/util"
	"github.com/dgryski/dgoogauth"
	"image"
	"net/url"
	"rsc.io/qr"
)

const (
	qrIssuer = "AuthService"
)

func GenerateQRCode(email string) (image.Image, string, error) {
	secret := util.GenerateRandomString(10, util.CharSet)
	encodedSecret := base32.StdEncoding.EncodeToString([]byte(secret))
	URL, err := url.Parse("otpauth://totp")
	if err != nil {
		return nil, "", err
	}

	URL.Path = fmt.Sprintf("%s/%s:%s", URL.Path, url.PathEscape(qrIssuer), url.PathEscape(email))

	params := url.Values{}
	params.Add("secret", encodedSecret)
	params.Add("issuer", qrIssuer)

	URL.RawQuery = params.Encode()
	fmt.Printf("URL is: %s\n", URL.String())

	code, err := qr.Encode(URL.String(), qr.Q)
	if err != nil {
		return nil, "", err
	}

	// PNG() returns a PNG image displaying the code.
	b := code.PNG()
	img, err := createImage(b)
	if err != nil {
		return nil, "", err
	}

	return img, secret, nil
}

func ValidateQRCode(code string, secret string) (bool, error) {
	encodedSecret := base32.StdEncoding.EncodeToString([]byte(secret))

	otpc := &dgoogauth.OTPConfig{
		Secret:      encodedSecret,
		WindowSize:  0,
		HotpCounter: 0,
	}

	authenticated, err := otpc.Authenticate(code)
	if err != nil {
		return false, err
	}

	return authenticated, nil
}

func createImage(b []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	return img, err
}
