package serverx

type testID string

func (ti testID) String() string { return string(ti) }

var (
	idA = testID("A")
	idB = testID("B")
	idC = testID("C")
	idD = testID("D")
)
