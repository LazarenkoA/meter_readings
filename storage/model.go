package storage

type IStorage interface {
	StoreObject(name string, object any) error
	RestoreObject(name string) (object map[string]interface{}, err error)
	DeleteObject(name string) error
}
