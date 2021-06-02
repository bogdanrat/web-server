package lib

import "regexp"

const (
	ImageExtensionPattern string = "([.|\\w|\\s|-])*\\.(?:jpg|gif|png)"
)

var (
	ImageRegex = regexp.MustCompile(ImageExtensionPattern)
)

func IsImage(value string) bool {
	return ImageRegex.MatchString(value)
}
