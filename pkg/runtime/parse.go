package runtime

import (
	"errors"
	"strings"
)

func allRuntimes() []*Runtime {
	return []*Runtime{
		{
			GOOS:   "darwin",
			GOARCH: "amd64",
		},
		{
			GOOS:   "darwin",
			GOARCH: "arm64",
		},
		{
			GOOS:   "linux",
			GOARCH: "amd64",
		},
		{
			GOOS:   "linux",
			GOARCH: "arm64",
		},
		{
			GOOS:   "windows",
			GOARCH: "amd64",
		},
		{
			GOOS:   "windows",
			GOARCH: "arm64",
		},
	}
}

func GetRuntimes(env string) ([]*Runtime, error) {
	if env == "all" {
		return allRuntimes(), nil
	}
	o, a, f := strings.Cut(env, "/")
	if f {
		return []*Runtime{
			{
				GOOS:   o,
				GOARCH: a,
			},
		}, nil
	}
	switch o {
	case "darwin", "linux", "windows":
		return []*Runtime{
			{
				GOOS:   o,
				GOARCH: "amd64",
			},
			{
				GOOS:   o,
				GOARCH: "arm64",
			},
		}, nil
	case "amd64", "arm64":
		return []*Runtime{
			{
				GOOS:   "darwin",
				GOARCH: o,
			},
			{
				GOOS:   "windows",
				GOARCH: o,
			},
			{
				GOOS:   "linux",
				GOARCH: o,
			},
		}, nil
	}
	return nil, errors.New("unsupported runtime")
}

const numOfAllRuntimes = 6

func GetRuntimesFromEnvs(envs []string) ([]*Runtime, error) {
	if envs == nil {
		return allRuntimes(), nil
	}
	ids := make(map[string]struct{}, numOfAllRuntimes)
	ret := make([]*Runtime, 0, numOfAllRuntimes)
	for _, env := range envs {
		rts, err := GetRuntimes(env)
		if err != nil {
			return nil, err
		}
		if len(rts) == numOfAllRuntimes {
			return rts, nil
		}
		for _, rt := range rts {
			id := rt.Env()
			if _, ok := ids[id]; ok {
				continue
			}
			ids[id] = struct{}{}
			ret = append(ret, rt)
		}
	}
	return ret, nil
}
