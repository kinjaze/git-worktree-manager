package metadata

type Store interface {
	Load() (File, error)
	Save(File) error
	Upsert(Record) error
	Remove(id string) error
}
