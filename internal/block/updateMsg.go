package block

type Updates struct {
	Updates Update       `json:"updates"`
	Deletes DeleteUpdate `json:"deletes"`
}

type Update struct {
	Updates map[int64][]*Block `json:"updates"`
}

type DeleteUpdate struct {
	NumClients    int64           `json:"numClients"`
	ClientDeletes []ClientDeletes `json:"clientDeletes"`
}

type ClientDeletes struct {
	Client        int64         `json:"client"`
	DeletedRanges []DeleteRange `json:"deletedRanges"`
}

type DeleteRange struct {
	StartClock   int64 `json:"startClock"`
	DeleteLength int64 `json:"deleteLength"`
}
