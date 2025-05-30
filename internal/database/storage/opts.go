package storage

type Option func(*Storage)

func WithWAL(wal WAL) Option {
	return func(storage *Storage) {
		storage.wal = wal
	}
}

func WithReplication(replication Replication) Option {
	return func(storage *Storage) {
		storage.replication = replication
	}
}
