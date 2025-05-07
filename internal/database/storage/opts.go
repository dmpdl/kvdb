package storage

type Option func(*Storage)

func WithWAL(wal WAL) Option {
	return func(storage *Storage) {
		storage.wal = wal
	}
}
