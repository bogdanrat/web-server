package lib

import "regexp"

const (
	ImageExtensionPattern string = "([.|\\w|\\s|-])*\\.(?:jpg|gif|png|JPG|GIF|PNG)"
)

var (
	ImageRegex = regexp.MustCompile(ImageExtensionPattern)
)

func IsImage(value string) bool {
	return ImageRegex.MatchString(value)
}
