package version

import (
	"fmt"
)

const (
	Major = "0"
	Minor = "0"
	Build = "1"
)

func Full() string {
	return fmt.Sprintf("%s.%s.%s", Major, Minor, Build)
}
