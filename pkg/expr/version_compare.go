package expr

import (
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

type Compare struct {
	op      string
	compare func(*version.Version) bool
}

func comparisons(sv1 *version.Version) []*Compare {
	return []*Compare{
		{
			op:      ">=",
			compare: sv1.GreaterThanOrEqual,
		},
		{
			op:      "<=",
			compare: sv1.LessThanOrEqual,
		},
		{
			op: "!=",
			compare: func(v *version.Version) bool {
				return !sv1.Equal(v)
			},
		},
		{
			op:      ">",
			compare: sv1.GreaterThan,
		},
		{
			op:      "<",
			compare: sv1.LessThan,
		},
		{
			op:      "=",
			compare: sv1.Equal,
		},
	}
}

func getCompareFunc(v string) func(s string) bool {
	return func(s string) bool {
		return compare(s, v)
	}
}

var commitHash = regexp.MustCompile(`^[0-9a-f]{40}$`)

func compare(constr, ver string) bool {
	if commitHash.MatchString(ver) {
		return false
	}
	sv1, err := version.NewVersion(ver)
	if err != nil {
		panic(err)
	}
	for _, constraint := range strings.Split(strings.TrimSpace(constr), ",") {
		c := strings.TrimSpace(constraint)
		matched := false
		for _, comp := range comparisons(sv1) {
			s := strings.TrimPrefix(c, comp.op)
			if s == c {
				continue
			}
			sv2, err := version.NewVersion(strings.TrimSpace(s))
			if err != nil {
				panic(err)
			}
			if !comp.compare(sv2) {
				return false
			}
			matched = true
			break
		}
		if !matched {
			panic("invalid operator. Operator must be one of >=, >, <, <=, !=, =")
		}
	}
	return true
}
