package pkg

type ID struct {
	ClientID int64
	Clock    int64
}

type StateVector struct {
	store map[int64]ID
}
