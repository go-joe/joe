+++
title = "Implement a new Memory"
slug = "memory"
weight = 6
+++

Memory modules let you persist key value data so it can be accessed again later.
Joe currently has the following five Memory implementations:

- In-Memory: https://github.com/go-joe/joe
- Redis Memory: https://github.com/go-joe/redis-memory
- File Memory: https://github.com/go-joe/file-memory
- Bolt Memory: https://github.com/robertgzr/joe-bolt-memory
- SQLite Memory: https://github.com/warmans/sqlite-memory

If you want to use some other system or technology to let your bot to persist
records, you can write your own Memory implementation.

### Memories are Modules

Firstly, your memory should be available as [`joe.Module`][module] so it can
easily be integrated into the bot via the [`joe.New(…)`][new] function.

The `Module` interface looks like this:

```go
// A Module is an optional Bot extension that can add new capabilities such as
// a different Memory implementation or Adapter.
type Module interface {
	Apply(*Config) error
}
```

To easily implement a Module without having to declare an `Apply` function on
your Memory type, you can use the `joe.ModuleFunc` type. For instance the
Redis memory uses the following, to implement it's `Memory(…)` function:

```go
// Memory returns a joe Module that configures the bot to use Redis as key-value
// store.
func Memory(addr string, opts ...Option) joe.Module {
	return joe.ModuleFunc(func(joeConf *joe.Config) error {
		conf := Config{Addr: addr}
		for _, opt := range opts {
			err := opt(&conf)
			if err != nil {
				return err
			}
		}

		if conf.Logger == nil {
			conf.Logger = joeConf.Logger("redis")
		}

		memory, err := NewMemory(conf)
		if err != nil {
			return err
		}

		joeConf.SetMemory(memory)
		return nil
	})
}
```

The passed `*joe.Config` parameter can be used to lookup general options such as
the `context.Context` used by the bot. Additionally you can create a named
logger via the `Config.Logger(…)` function.
 
Most importantly for a Memory implementation however is, that it finally needs
to register itself via the `Config.SetMemory(…)` function.

By defining a `Memory(…)` function in your package, it is now possible to use
your memory as Module passed to `joe.New(…)`. Additionally your `NewMemory(…)`
function is useful to directly create a new memory instance which can be used
during unit tests. Last but not least, the options pattern has proven useful in
this kind of setup and is considered good practice when writing modules in general.

### The Memory Interface

```go
// The Memory interface allows the bot to persist data as key-value pairs.
// The default implementation of the Memory is to store all keys and values in
// a map (i.e. in-memory). Other implementations typically offer actual long term
// persistence into a file or to Redis.
type Memory interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, bool, error)
	Delete(key string) (bool, error)
	Keys() ([]string, error)
	Close() error
}
```

Looking at the interface you can see that the Memory must implement all CRUD
operations (Create, Read, Update & Delete) as well as a function to retrieve all
previously stored keys and a function to close the connection and release any
held resources.

### Storage encoding

Each Memory implementation manages key value data, where the keys are strings
and the values are only bytes. In the event handlers, the memory can be
accessed via the bots concrete `Storage` type which accepts values as interfaces
and provides read access by unmarshalling values back into types via a pointer,
very much like you may know already from Go's standard library (e.g. `encoding/json`).

To encode the given `interface{}` values into the `[]byte` that is passed to your
Memory implementation, the storage also has a `MemoryEncoder` which is defined as:

```go
// A MemoryEncoder is used to encode and decode any values that are stored in
// the Memory. The default implementation that is used by the Storage uses a
// JSON encoding.
type MemoryEncoder interface {
	Encode(value interface{}) ([]byte, error)
	Decode(data []byte, target interface{}) error
}
``` 

If you want, you can change the encoding from JSON to something else (e.g. to
implement encryption) by providing a type that implements this interface and
then using the `joeConf.SetMemoryEncoder(…)` function in your Module during the setup.

### Getting Help

Generally writing a new Memory implementation should not be very hard but it's a
good idea to look at the other Memory implementations to get a better
understanding of how to implement your own. If you have questions or need help,
simply open an issue at the [Joe repository at GitHub](https://github.com/go-joe/joe/issues/new).  

Happy coding :robot:

[module]: https://godoc.org/github.com/go-joe/joe#Module
[new]: https://godoc.org/github.com/go-joe/joe#New
