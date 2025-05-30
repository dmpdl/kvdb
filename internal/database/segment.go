package database

type Segment struct {
	Name string
	Logs []WALRecord
}
