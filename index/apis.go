package main

import(
)

type BlobStore interface{
	Store(ctx context.Context, data []byte, group string, timestamp time.Time, processed bool) (id string, error)
	List(ctx context.Context, group string, startingTime time.Time, maxTime time.Time, maxCount int) (ids []string, error)
	Delete(ctx context.Context, ids []string) error
	GetBytes(ctx context.Context, requests ByteRequest) 
	Freeeze(ctx context.Context, ids []string, callback func(error))
	Thaw(ctx context.Context, ids []string, callback func(error))
	Lock(ctx context.Context, ids []string) (lockedIndicator []bool, unlock func(), err error)
	MarkProcessed(ctx context.Context, ids []string) error
	FindUnlockedUnprocessed(ctx context.Context, limit int) ([]string, error)
}

type AttributeStore interface{
	Lookup(ctx context.Context, name string) Attribute

type Attribute struct {
	Name	string
	Number	uint32
	Settings	
}

type Estimator interface{
	See(key, value)
	Forget(key, value)
	TotalSeen() uint64
	Estimate(key, value) uint64
}

type MetaKVStore interface{
	Store(key, value, esitmate, timestamp, docid)
	Forget(key, value, timestamp, docid)
	Find(key, value, timestamp range)
}

type KVStore interface{
	Store(key, value, timestamp, docid)
	Forget(key, value, timestamp, docid)
	Find(key, value, timestamp range)
}

type AttributeTracker interface{
	GetAttributeByName(context.Context, name string) AttribueSettings
	UpdateAttribute(context.Context, string, AttributeSettings) error
}

type AttributeSettings struct {
	Number	int
}



type Component[I interface{}] 

type ComponentRegistry[I interface] 

