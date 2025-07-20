package storage

type IStorage interface {
	StoreObject(name string, object any) error
	RestoreObject(name string) (object map[string]interface{}, err error)
	RestoreAsObject(name string, callback func(data []byte) error) error
	DeleteObject(name string) error
}
