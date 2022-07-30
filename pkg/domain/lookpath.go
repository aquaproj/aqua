package domain

type ExecFinder interface {
	LookPath(string) (string, error)
}
