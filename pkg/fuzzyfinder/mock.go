package fuzzyfinder

type MockFuzzyFinder struct {
	idxs []int
	err  error
}

func NewMock(idxs []int, err error) *MockFuzzyFinder {
	return &MockFuzzyFinder{
		idxs: idxs,
		err:  err,
	}
}

func (f *MockFuzzyFinder) Find(items []Item, hasPreview bool) (int, error) {
	return f.idxs[0], f.err
}

func (f *MockFuzzyFinder) FindMulti(items []Item, hasPreview bool) ([]int, error) {
	return f.idxs, f.err
}
