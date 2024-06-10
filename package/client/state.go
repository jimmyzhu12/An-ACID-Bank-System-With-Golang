package client

type Client struct {
	Id        string
	branches  map[BranchID]*Branch
	cmdCh     chan Command
	coordBr   Branch
	Timestamp string
}
