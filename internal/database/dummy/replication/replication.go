package replication

import "context"

type Replication struct{}

func New() *Replication {
	return &Replication{}
}

func (r Replication) Run(_ context.Context) {}
