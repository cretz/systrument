package util

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"strings"
)

type Versionable interface {
	Version() (*version.Version, error)
}

func CheckVersion(v Versionable, name string, c version.Constraints) error {
	if vers, err := v.Version(); err != nil {
		return fmt.Errorf("Unable to use %v: %v", name, err)
	} else if !c.Check(vers) {
		return fmt.Errorf("Invalid version for %v. Expected within %v, got %v", name, c, vers)
	}
	return nil
}

func Constraint(str string) version.Constraints {
	c, err := version.NewConstraint(str)
	if err != nil {
		panic(err)
	}
	return c
}

// Gets version after last space and trims off anything anything non-numeric after the second decimal
// (so it will get rid of metadata)
func VersionAfterLastSpace(str string) (*version.Version, error) {
	segments := strings.SplitN(str[strings.LastIndex(str, " ")+1:], ".", 3)
	for i, v := range segments[len(segments)-1] {
		if v < '0' || v > '9' {
			segments[len(segments)-1] = segments[len(segments)-1][0:i]
			break
		}
	}
	return version.NewVersion(strings.Join(segments, "."))
}
