/*
 * Copyright (c) 2022, 2025 Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * https://oss.oracle.com/licenses/upl.
 */

/*
Package coherence provides a set of functions and interfaces for Go programs to act as cache clients to a
Coherence Cluster using gRPC for the network transport.

Your cluster must be running Coherence Community Edition (CE) 22.06.4+ or Coherence commercial
version 14.1.1.2206.4+ and must be running a gRPC Proxy.

Two interfaces, [NamedMap] and [NamedCache], are available to access Coherence caches. [NamedCache] is syntactically identical in behaviour to a [NamedMap],
but additionally implements the PutWithExpiry operation.

# Introduction

The Coherence Go client provides the following features:

  - Familiar Map-like interface for manipulating cache entries including but not limited to Put, PutWithExpiry, PutIfAbsent, PutAll, Get, GetAll, Remove, Clear, GetOrDefault, Replace, ReplaceMapping, Size, IsEmpty, ContainsKey, ContainsValue, ContainsEntry
  - Cluster-side querying, aggregation and filtering of map entries
  - Cluster-side manipulation of map entries using EntryProcessors
  - Registration of listeners to be notified of mutations such as
  - insert, update and delete on Maps, map lifecycle events such as truncated, released or destroyed
    and session lifecycle events such as connected, disconnected, reconnected and closed
  - Support for storing Go structs as JSON as well as the ability to serialize to Java objects on the server for access from other Coherence language API's
  - Near cache support to cache frequently accessed data in the Go client to avoid sending requests across the network
  - Support for simple and double-ended queues in Coherence Community Edition 24.09+ and commercial version 14.1.2.0+
  - Full support for Go generics in all Coherence API's

For more information on Coherence caches, please see the [Coherence Documentation].

# Supported Go versions

This API fully supports Go Generics and is only supported for use with Go versions 1.23 and above.

# Obtaining a Session

Example:

	import (
	    coherence "github.com/oracle/coherence-go-client/v2/coherence"
	)

	...

	session, err := coherence.NewSession(ctx)
	if err != nil {
	    log.Fatal(err)
	}
	defer session.Close()

The [NewSession] function creates a new session that will connect to a gRPC proxy server on "localhost:1408" by default.

You can specify the host and port to connect to by specifying the environment variable COHERENCE_SERVER_ADDRESS.
See [gRPC Naming] for information on values for this.

You can also pass [coherence.WithAddress]("host:port") to specify the gRPC host and port to connect to. The default
connection mode is with SSL enabled, but you can use plan-text via using [coherence.WithPlainText]().

	session, err := coherence.NewSession(ctx, coherence.WithPlainText(), coherence.WithAddress("my-host:7574"))

You are also able to use the 'coherence' gRPC resolver address of "coherence:///host:port"
to connect to the Coherence Name Service, running on the cluster port, and automatically discover the gRPC endpoints. For example:

	session, err := coherence.NewSession(ctx, coherence.WithPlainText(), coherence.WithAddress("coherence:///localhost:7574"))

If you have multiple clusters on the same port, you can also append the cluster name to specify which cluster you wish to contact.

	coherence.WithAddress("coherence:///localhost:7574/cluster2")

When using the 'coherence' gRPC resolver, the Go client randomizes any addresses that are returned to help
load balance across gRPC proxies. You can turn this off by setting the environment variable COHERENCE_RESOLVER_RANDOMIZER=false.

To Configure SSL, you must first enable SSL on the gRPC Proxy, see [gRPC Proxy documentation] for details.
Refer to the section on [NewSession] for more information on setting up a SSL connection on the client.

See [SessionOptions] which lists all the options supported by the [Session] API.

# Controlling timeouts

Most operations you call require you to supply a [context.Context]. If your context does not contain a deadline,
the operation will wrap your context in a new [context.WithTimeout] using either the default timeout of 30,000 millis or
the value you set using option [coherence.WithRequestTimeout] when you called [NewSession].

For example, to override the default request timeout of 30,000 millis with one of 5 seconds for a [Session] you can do the following:

	session, err = coherence.NewSession(ctx, coherence.WithRequestTimeout(time.Duration(5) * time.Second))

You can also override the default request timeout using the environment variable COHERENCE_CLIENT_REQUEST_TIMEOUT.

By default, if an endpoint is not ready, the Go client will fail-fast. You can change this behaviour by setting
the option [coherence.WithReadyTimeout] to a value millis value greater than zero which will cause the Go client
to wait until up to the timeout specified until it fails if no endpoint is available. You can also use the environment variable
COHERENCE_READY_TIMEOUT.

You also have the ability to control maximum amount of time, in milliseconds, a [Session] may remain in a disconnected state
without successfully reconnecting. For this you use the option [coherence.WithDisconnectTimeout] or the environment
variable COHERENCE_SESSION_DISCONNECT_TIMEOUT.

# Setting Log Levels

The Coherence Go client supports setting the following log levels to change verbosity of messages output.
The default level is 'INFO' and this can be changed by setting the environment variable COHERENCE_LOG_LEVEL to one of the following values:

- ERROR

- WARNING

- INFO

- DEBUG

- ALL

All messages at and above the level are displayed. For example setting to WARNING will only show ERROR and WARNING messages.
Where as setting DEBUG, will show ERROR, WARNING, INFO and DEBUG messages.

Note: Setting to ALL should only be used to diagnose issues as directed by Oracle Support. This level will
output a large volume of messages.

# Obtaining a NamedMap or NamedCache

Once a session has been created, the [GetNamedMap](session, name, ...options) or [GetNamedCache](session, name, ...options)
can be used to obtain an instance of a [NamedMap] or [NamedCache]. The key and value types must be provided as generic type arguments.
This identifier may be shared across clients.  It's also possible to have many [NamedMap]s or [NamedCache]s defined and in use simultaneously.

Example:

	session, err := coherence.NewSession(ctx)
	if err != nil {
	    log.Fatal(err)
	}
	defer session.Close()

	namedMap, err := coherence.GetNamedMap[int, string](session, "customers")
	if err != nil {
	    log.Fatal(err)
	}

If you wish to create a [NamedCache], which supports expiry, you can use the [GetNamedCache] function and then use the PutWithExpiry function call.

	namedCache, err := coherence.GetNamedCache[int, string](session, "customers")
	if err != nil {
	    log.Fatal(err)
	}

	_, err = namedCache.PutWithExpiry(ctx, person1.ID, person1, time.Duration(5)*time.Second)

If your [NamedCache] requires the same expiry for every entry, you can use the [coherence.WithExpiry] cache option.
Each call to Put will use the default expiry you have specified. If you use PutWithExpiry, this will override the default
expiry for that key.

	namedCache, err := coherence.GetNamedCache[int, Person](session, "cache-expiry", coherence.WithExpiry(time.Duration(5)*time.Second))

See [CacheOptions] which lists all the options supported by the [GetNamedCache] or [GetNamedMap] API.

# Basic CRUD operations

Note: See the [examples] on GitHub for detailed examples.

Assuming a very trivial [NamedMap] with integer keys and string values.

	session, err := coherence.NewSession(coherence.WithPlainText())
	if err != nil {
	    log.Fatal(err)
	}

	namedMap, err := coherence.GetNamedMap[int, string](session, "my-map")
	if err != nil {
	    log.Fatal(err)
	}

	ctx := context.Background()

	// put a new key / value
	if _, err = namedMap.Put(ctx, 1, "one"); err != nil {
	    log.Fatal(err)
	}

	// get the value for the given key
	if value, err = namedMap.Get(ctx, 1); err != nil {
	    log.Fatal(err)
	}
	fmt.Println("Value for key 1 is", *value)

	// update the value for key 1
	if _, err = namedMap.Put(ctx, 1, "ONE"); err != nil {
	    log.Fatal(err)
	}

	// retrieve the updated value for the given key
	if value, err = namedMap.Get(ctx, 1); err != nil {
	    log.Fatal(err)
	}
	fmt.Println("Updated value is", *value)

	if _, err = namedMap.Remove(ctx, 1); err != nil {
	    log.Fatal(err)
	}

Note: Keys and values are serialized to JSON and stored in Coherence as a com.oracle.coherence.io.json.JsonObject.
if you wish to store structs as native Java objects, then please see the section further down on "Serializing to Java Objects on the Server".

# Working with structs

	type Person struct {
	    ID   int    `json:"id"`
	    Name string `json:"name"`
	    Age  int    `json:"age"`
	}

	// create a new NamedMap of Person with key int
	namedMap, err := coherence.GetNamedMap[int, Person](session, "test")
	if err != nil {
	    log.Fatal(err)
	}

	// clear the Map
	if err = namedMap.Clear(ctx); err != nil {
	    log.Fatal(err)
	}

	newPerson := Person{ID: 1, Name: "Tim", Age: 21}
	fmt.Println("Add new Person", newPerson)
	if _, err = namedMap.Put(ctx, newPerson.Id, newPerson); err != nil {
	    log.Fatal(err)
	}

	// retrieve the Person
	if person, err = namedMap.Get(ctx, 1); err != nil {
	    log.Fatal(err)
	}
	fmt.Println("Person from Get() is", *person)

	// Update the age using and entry processor for in-place processing
	_, err = coherence.Invoke[int, Person, bool](ctx, namedMap, 1, processors.Update("age", 56))
	if err != nil {
	    log.Fatal(err)
	}

	// retrieve the updatedPerson
	if person, err = namedMap.Get(ctx, 1); err != nil {
	    log.Fatal(err)
	}
	fmt.Println("Person is", *person)

# Querying and filtering using channels

Channels are used to deal with individual keys, values or entries
streamed from the backend using a filter or an open query.  Depending
upon the operation, each result element is wrapped in one of the structs
[StreamedEntry], [StreamedValue] or [StreamedKey] which wraps an error and a
Key and/or a Value. As always, the Err object must be checked for errors before accessing the Key or Value fields.
All functions that return channels are EntrySetFilter, KeySetFilter, ValuesFilter,
EntrySet, KeySet, Values, InvokeAll and InvokeAllFilter.

	namedMap, err := coherence.GetNamedMap[int, Person](session, "people")
	if err != nil {
	    log.Fatal(err)
	}

	// extractors
	age := extractors.Extract[int]("age")
	name := extractors.Extract[string]("name")

	// retrieve all people aged > 30
	ch := namedMap.EntrySetFilter(ctx, filters.Greater(age, 20))
	for result := range ch {
	    if result.Err != nil {
	        log.Fatal(result.Err)
	    }
	    fmt.Println("Key:", result.Key, "Value:", result.Value)
	}

	// we can also do more complex filtering such as looking for people > 30 and where there name begins with 'T'
	ch := namedMap.EntrySetFilter(ctx, filters.Greater(age, 20).And(filters.Like(name, "T%", true)))

If you want to sort the results from the EntrySetFilter command you can use the following function
[EntrySetFilterWithComparator]. Due generics limitations in Go, this is not a function call off the [NamedMap]
or [NamedCache] interface, but a function call that takes a [NamedMap] or [NamedCache].

	age := extractors.Extract[int]("age")

	fmt.Println("Retrieve the people between the age of 17 and 21 and order by age ascending")
	ch := coherence.EntrySetFilterWithComparator(ctx, namedMap, filters.Between(age, 17, 21), extractors.ExtractorComparator(age, true))
	for result := range ch {
	    if result.Err != nil {
	        panic(err)
	    }
	    fmt.Printf("Key: %v, Value: %s\n", result.Key, result.Value.String())
	}

Note: the entries are sorted internally on the gRPC proxy to avoid excessive memory usage, but you need to be careful
when running this operation against NamedCaches with large number of entries.

Sorting via a [Comparator] is only available when connecting to Coherence server versions CE 25.03+ and commercial 14.1.2.0+.

# Using entry processors for in-place processing

A Processor is an object that allows you to process (update) one or more [NamedMap] entries on the [NamedMap] itself,
instead of moving the entries to the client across the network. In other words, using processors we send
the processing to where the data resides thus avoiding massive data movement across the network. Processors can be
executed against all entries, a single key or against a set of entries that match a Filter.

To demonstrate this, lets assume we have a [NamedMap] populated with Person struct below, and we want to
run various scenarios to increase peoples salary by using a [processors.Multiply] processor.

	type Person struct {
	    Id     int     `json:"id"`
	    Name   string  `json:"name"`
	    Salary float32 `json:"salary"`
	    Age    int     `json:"age"`
	    City   string  `json:"city"`
	}

	namedMap, err := coherence.GetNamedMap[int, Person](session, "people")

	// 1. Increase the salary of the person with Id = 1
	newSalary, err = coherence.Invoke[int, Person, float32](ctx, namedMap, 1, processors.Multiply("salary", 1.1, true))

	city := extractors.Extract[string]("city")

	// 2. Increase the salary of all people in Perth
	ch2 := coherence.InvokeAllFilter[int, Person, float32](ctx, namedMap, filters.Equal(city, "Perth"), processors.Multiply("salary", 1.1, true))
	for result := range ch2 {
	    if result.Err != nil {
	        log.Fatal(result.Err)
	    }
	}

	// 3. Increase the salary of people with Id 1 and 5
	ch2 := coherence.InvokeAllKeys[int, Person, float32](ctx, namedMap, []int{1, 5}, processors.Multiply("salary", 1.1, true))
	for result := range ch2 {
	    if result.Err != nil {
	        log.Fatal(result.Err)
	    }
	}

# Aggregating cache data

Aggregators can be used to perform operations against a subset of entries to obtain a single result.
Entry aggregation occurs in parallel across the grid to provide map-reduce support when working with
large amounts of data.

To demonstrate this, lets assume we have a [NamedMap] populated with Person struct as per the previous example, and we want to
run various scenarios to perform aggregations.

	namedMap, err := coherence.GetNamedMap[int, Person](session, "people")
	if err != nil {
	    log.Fatal(err)
	}

	// Retrieve the distinct cities from all people
	citiesValues, err := coherence.Aggregate(ctx, namedMap, extractors.Extract[string]("city"))
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Println(*citiesValues)
	// output: [Perth, Melbourne, Brisbane]

	age := extractors.Extract[int]("age")

	// minimum age across keys 3 and 4
	ageResult, err = coherence.AggregateKeys(ctx, namedMap, []int{3, 4}, aggregators.Min(age))

	// top 2 people by salary using filter
	var salaryResult *[]Person
	salaryResult, err = coherence.AggregateFilter[int, Person, []Person](ctx, namedMap, filters.Greater(age, 40),
	    aggregators.TopN[float32, Person](extractors.Extract[float32]("salary"), false, 2))

# Responding to cache events

The Coherence Go client provides the ability to add a [MapListener] that will receive events (inserts, updates, deletes)
that occur against a [NamedMap] or [NamedCache]. You can listen for all events, events based upon a filter or
vents based upon a key.

	// in your main code, create a new NamedMap and register the listener
	namedMap, err := coherence.GetNamedMap[int, Person](session, "people")
	if err != nil {
	    log.Fatal(err)
	}

	listener := coherence.NewMapListener[int, Person]().OnUpdated(
	func(e coherence.MapEvent[int, Person]) {
	    key, err := e.Key()
	    if err != nil {
	        panic("unable to deserialize key")
	    }

	    newValue, err := e.NewValue()
	    if err != nil {
	        panic("unable to deserialize new value")
	    }

	    oldValue, err := e.OldValue()
	    if err != nil {
	        panic("unable to deserialize old value")
	    }

	    fmt.Printf("**EVENT=Updated: key=%v, oldValue=%v, newValue=%v\n", *key, *oldValue, *newValue)
	})

	if err = namedMap.AddListener(ctx, listener); err != nil {
	    panic(err)
	}

	// ensure we unregister the listener
	defer func(ctx context.Context, namedMap coherence.NamedMap[int, Person], listener coherence.MapListener[int, Person]) {
	    _ = namedMap.RemoveListener(ctx, listener)
	}(ctx, namedMap, listener)

	// As you carry out operations that will mutate the cache entries, update the age to 56, you will see the events printed
	_, err = coherence.Invoke[int, Person, bool](ctx, namedMap, 1, processors.Update("age", 56))
	if err != nil {
	    log.Fatal(err)
	}

	// output:
	// **EVENT=Updated: key=1, oldValue={1 Tim 53}, newValue={1 Tim 53}

	// you can also listen based upon filters, for example the following would create a
	// listener for all entries where the salary is > 17000
	if err = namedMap.AddFilterListener(ctx, listener,
	    filters.Greater(extractors.Extract[int]("salary"), 17000)); err != nil {
	    log.Fatal("unable to add listener", listener, err)
	}

	// You can also listen on a specific key, e.g. list on key 1.
	listener := NewUpdateEventsListener[int, Person]()
	if err = namedMap.AddKeyListener(ctx, listener, 1); err != nil {
	    log.Fatal("unable to add listener", listener, err)
	}

# Responding to cache lifecycle events

The Coherence Go client provides the ability to add a [MapLifecycleListener] that will receive events (truncated and destroyed)
that occur against a [NamedMap] or [NamedCache].

	// consider the example below where we want to listen for all 'truncate' events for a NamedMap.
	// in your main code, create a new NamedMap and register the listener
	namedMap, err := coherence.GetNamedMap[int, Person](session, "people")
	if err != nil {
	    log.Fatal(err)
	}

	// Create a listener and add to the cache
	listener := coherence.NewMapLifecycleListener[int, Person]().
	    OnTruncated(func(e coherence.MapLifecycleEvent[int, Person]) {
	        fmt.Printf("**EVENT=%s: source=%v\n", e.Type(), e.Source())
	    })

	namedMap.AddLifecycleListener(listener)
	defer namedMap.RemoveLifecycleListener(listener)

	newPerson := Person{ID: 1, Name: "Tim", Age: 21}
	fmt.Println("Add new Person", newPerson)
	if _, err = namedMap.Put(ctx, newPerson.Id, newPerson); err != nil {
	    log.Fatal(err)
	}

	if size, err = namedMap.Size(ctx); err != nil {
	    log.Fatal(err)
	}
	fmt.Println("Cache size is", size, "truncating cache")

	if err = namedMap.Truncate(ctx); err != nil {
	    log.Fatal(err)
	}

	time.Sleep(time.Duration(5) * time.Second)

	// output:
	// Add new Person {1 Tim 53}
	// Cache size is 1 truncating cache
	// **EVENT=Truncated: value=NamedMap{name=people, format=json}

# Responding to session lifecycle events

The Coherence Go client provides the ability to add a [SessionLifecycleListener] that will receive events (connected, closed,
disconnected or reconnected) that occur against the [Session].
Note: These events use and experimental gRPC API so may not be reliable or may change in the future. This is due to the
experimental nature of the underlying gRPC API.

Consider the example below where we want to listen for all 'All' events for a [Session].
in your main code, create a new [Session] and register the listener

	// create a new Session
	session, err := coherence.NewSession(ctx, coherence.WithPlainText())
	if err != nil {
	    log.Fatal(err)
	}

	// Create a listener to listen for session events
	listener := coherence.NewSessionLifecycleListener().
	    OnAny(func(e coherence.SessionLifecycleEvent) {
	        fmt.Printf("**EVENT=%s: source=%v\n", e.Type(), e.Source())
	})

	session.AddSessionLifecycleListener(listener)
	defer session.RemoveSessionLifecycleListener(listener)

	// create a new NamedMap of Person with key int
	namedMap, err := coherence.GetNamedMap[int, Person](session, "people")
	if err != nil {
	    log.Fatal(err)
	}

	// clear the Map
	if err = namedMap.Clear(ctx); err != nil {
	    log.Fatal(err)
	}

	session.Close()

	time.Sleep(time.Duration(5) * time.Second)

	// output:
	// 2023/01/31 11:15:37 connected session 59f3ec81-dda1-41b7-92de-70aad3d26615 to address localhost:1408
	// 2023/01/31 11:15:38 closed session 59f3ec81-dda1-41b7-92de-70aad3d26615
	// **EVENT=session_closed: source=SessionID=59f3ec81-dda1-41b7-92de-70aad3d26615, closed=true, caches=0, maps=0

# Working with Queues

When connecting to a Coherence CE cluster versions 24.09 or above or commercial 14.1.2.0.+, you have the ability to create two main types of queues, a [NamedQueue] or [NamedDequeue].

A [NamedQueue] is a simple FIFO queue which can be one of two types: either [Queue] - a simple queue which stores data in a single
partition and is limited to approx 2GB of storage, or [PagedQueue] which distributes data over the cluster and is only limited
by the cluster capacity.

A [NamedDequeue] is a simple double-ended queue that stores data in a single partition.

Queues in general have the following methods. See [NamedQueue] for the full list.

- PeekHead(ctx context.Context) (*V, error) - retrieve but not remove the value at the head of this queue

- PollHead(ctx context.Context) (*V, error - retrieves and removes the head of this queue

- OfferTail(ctx context.Context, value V) error - inserts the specified value to the end of this queue if it is possible to do so

Consider the example below where we want to create a standard queue and add 10 entries, and then retrieve 10 entries.
We have specified [coherence.Queue] as the type but this could also be [coherence.PagedQueue].

	namedQueue, err := coherence.GetNamedQueue[string](ctx, session, "my-queue", coherence.Queue)
	if err != nil {
	    panic(err)
	}

	// add an entry to the tail of the queue
	for i := 1; i <= iterations; i++ {
	    v := fmt.Sprintf("value-%v", i)
	    log.Printf("OfferTail() %s to the queue\n", v)
	    err = namedQueue.OfferTail(ctx, v)
	    if err != nil {
	        panic(err)
	    }
	}
	// output:
	// Offer() value-1 to the queue
	// ...
	// Offer() value-10 to the queue

	// Poll 10 entries from the head of the queue
	for i := 1; i <= iterations; i++ {
	    value, err = namedQueue.PollHead(ctx)
	    if err != nil {
	        panic(err)
	    }
	    log.Printf("Poll() returned: %s\n", *value)
	}

	// output:
	// Poll() returned: value-1
	// ...
	// Poll() returned: value-10

	// try to read again should get nil as nothing left on the queue
	value, err = namedQueue.PollHead(ctx)
	if err != nil {
	    panic(err)
	}
	log.Println("last value is", value)
	// output: last value is nil

The [NamedDequeue] is a double-ended queue and has the following additional functions:

- OfferHead(ctx context.Context, value V) error - inserts the specific value at the head of this queue

- PollTail(ctx context.Context) (*V, error) - retrieves and removes the tail of this queue

- PeekTail(ctx context.Context) (*V, error) - retrieves, but does not remove, the tail of this queue

In the following example, we are using a [NamedDequeue] or double-ended queue, where we have the ability to
add or offer data to the head of the queue as well as the end of the queue, and also poll and peek the the end of the queue.

	namedQueue, err := coherence.GetNamedDeQueue[string](ctx, session, "double-ended-queue")
	if err != nil {
	    panic(err)
	}

	// add 10 entries to the end (tail) of the queue
	for i := 1; i <= iterations; i++ {
	    v := fmt.Sprintf("value-%v", i)
	    log.Printf("OfferTail() %s to the queue\n", v)
	    err = namedQueue.OfferTail(ctx, v)
	    if err != nil {
	        panic(err)
	    }
	}

	// output:
	// 2024/11/27 11:05:37 OfferTail() value-1 to the queue
	// ..
	// 2024/11/27 11:05:37 OfferTail() value-10 to the queue

	// Offer a value to the head
	err = namedQueue.OfferHead(ctx, "value-head")
	if err != nil {
	    panic(err)
	}

	// peek the tail of the queue
	value, err = namedQueue.PeekTail(ctx)
	if err != nil {
	    panic(err)
	}
	log.Printf("PeekTail() returned: %s\n", *value)

	// output:
	// 2024/11/27 11:05:37 PeekTail() returned: value-10

	// poll for iterations +1 because we added another entry to the head
	for i := 1; i <= iterations+1; i++ {
	    value, err = namedQueue.PollHead(ctx)
	    if err != nil {
	        panic(err)
	    }
	    log.Printf("PollHead() returned: %s\n", *value)
	}

	// output:
	// 2024/11/27 11:05:37 PollHead() returned: value-head (the value we added to the head)
	// 2024/11/27 11:05:37 PollHead() returned: value-1
	// ..
	// 2024/11/27 11:05:37 PollHead() returned: value-10

	// try to read again should get nil
	value, err = namedQueue.PollHead(ctx)
	if err != nil {
	    panic(err)
	}
	log.Println("last value is", value)

	// output:
	// 2024/11/27 11:05:37 last value is <nil>

See the [Queues] documentation for more information on using queues on the Coherence Server.

# Responding to queue lifecycle events

The Coherence Go client provides the ability to add a [QueueLifecycleListener] that will receive events (truncated, released and destroyed)
that occur against a [NamedQueue].

	// consider the example below where we want to listen for all 'QueueReleased' events for a NamedQueue.
	// in your main code, create a new NamedQueue and register the listener.
	// Note: this is a contrived example, but you can listen for QueueDestroyed and QueueTruncated events
	// in a similar way.

	namedQueue, err := coherence.GetNamedQueue[string](session, "queue", coherence.Queue)
	if err != nil {
	    log.Fatal(err)
	}

	// Create a listener to monitor
	listener := coherence.NewQueueLifecycleListener[string]().
	    OnTruncated(func(e coherence.QueueLifecycleEvent[string]) {
	        fmt.Printf("**EVENT=%s: source=%v\n", e.Type(), e.Source())
	    })

	_ = namedQueue.AddLifecycleListener(listener)
	defer namedQueue.RemoveLifecycleListener(listener)

	namedQueue.Release()

	// sleep to ensure we receive the event before we close
	time.Sleep(5 * time.Second)

	// output:
	// 2024/11/28 11:40:58 INFO: Session [b1435a16-f210-4289-97e4-e1654947acd5] connected to [localhost:1408] Coherence version: 24.09, serverProtocolVersion: 1, proxyMemberId: 1
	// **EVENT=queue_released: source=NamedQueue{name=queue-events, type=Queue, queueID=1198559040}
	// 2024/11/28 11:41:03 INFO: Session [b1435a16-f210-4289-97e4-e1654947acd5] closed

# Serializing to Java objects on the server

By default, the Coherence Go client serializes any keys and values to JSON and then stores them as JsonObjects in Coherence.
This is usually sufficient for most applications where you are only accessing your data via the Go client.

If you wish to access your data via other clients such as Java, JavaScript, C++, .NET or Python, it's best to use Java classes, known to Coherence server,
representing the data model. The following describes how to achieve interoperability with Java.

Step 1. Create your Java Classes

Firstly you must define your data model for all Java classes and configure for JSON serialization. You do not need to annotate all the attributes
with @JsonbProperty, but it is a good practice so that you have consistent names with Go. Below is a shorted version of a
Customer class without all the extras such as getters, setters, hashCode, etc, that you know you need. In the example below I am using
standard Java serialization, but you can use POF serialization if you have that configured.

	package com.oracle.demo;

	public class Customer implements Serializable {
		public Customer() {} // required

		@JsonbProperty("id")
		private int id;

		@JsonbProperty("customerName")
		private String customerName;

		@JsonbProperty("outstandingBalance")
		private double outstandingBalance;

		...

Step 2. Define your type alias.

In the code deployed to your Coherence storage-nodes, you need to create a file in your resources root called META-INF/type-aliases.properties
which contains an alias and fully qualified class name for each of your classes.

	# Example META-INF/type-aliases.properties file
	customer=com.oracle.demo.Customer
	order=com.oracle.demo.Order

Step 3. Define your Go structs

Next you need to define your Go structs with JSON names matching your Java objects. You also need to include a Class attribute with
the JSON attribute name of "@class". We will set this in our object to the value "customer" matching the value in the type-aliases.properties
on the server.

	type Customer struct {
	    Class              string   `json:"@class"`
	    ID                 int      `json:"id"`
	    CustomerName       string   `json:"customerName"`
	    OutstandingBalance float32  `json:"outstandingBalance"`
	}

Step 4. Create and put the value

Lastly, when you create a Customer object you must set the Class value matching the alias above.

	customer := Customer{
	    Class:              "customer",
	    ID:                 1,
	    CustomerName:       "Tim",
	    OutstandingBalance: 10000,
	}

	// store the entry in Coherence, it will be stored as a com.oracle.demo.Customer POJO!

	_, err = namedMap.Put(ctx, customer.ID, customer)
	if err != nil {
	    log.Fatal(err)
	}

# Using Near Caches

The Coherence Go client allows you to specify a near cache to cache frequently accessed data in your Go application.
When you access data using Get() or GetAll() operations, returned entries are stored in the near cache and subsequent data
access for keys in the near cache is almost instant where without a near cache each operation above always results in a network call.

On creating a near cache, Coherence automatically adds a [MapListener] to your [NamedMap] or [NamedCache] which listens on
all cache events and updates or invalidates entries in the near cache that have been changed or removed on the server.

To manage the amount of memory used by the near cache, the following options are supported when creating one:

  - time-to-live (TTL) – objects expired after time in near cache, e.g. 5 minutes
  - High-Units – maximum number of cache entries in the near cache
  - Memory – maximum amount of memory used by cache entries

Note: You can specify either High-Units or Memory and in either case, optionally, a TTL.

Note: The minimum expiry time for a near cache entry is 1/4 second. This is to ensure that expiry of elements is as efficient
as possible. You will receive an error if you try to set the TTL to a lower value.

The above can be specified by passing [NearCacheOptions] within [WithNearCache] when creating a [NamedMap] or [NamedCache].
See below for various ways of creating near caches.

You can ask a [NamedMap] or [NamedCache] for its near cache statistics by calling GetNearCacheStats(). Various statistics
are recorded in regard to the near cache and can be seen via the [CacheStats] interface. If the [NamedMap] or [NamedCache]
does not have a near cache, nil will be returned.

1. Creating a Near Cache specifying time-to-live (TTL)

The following example shows how to get a named cache that will cache entries from Get() or GetAll() for up to 30 seconds.

	// specify a TTL of 30 seconds
	nearCacheOptions := coherence.NearCacheOptions{TTL: time.Duration(30) * time.Second}

	namedMap, err := coherence.GetNamedMap[int, string](session, "customers", coherence.WithNearCache(&nearCacheOptions))
	if err != nil {
	    log.Fatal(err)
	}

	// issue first Get for data in the cache on the storage-nodes. Entries found will be stored in near cache
	value, err = namedMap.Get(ctx, 1)
	if err != nil {
	    panic(err)
	}

	// subsequent access will be almost instant from near cache
	value, err = namedMap.Get(ctx, 1)

	// you can check the near cache stats
	fmt.Println("Near cache size is", namedMap.GetNearCacheStats().Size())

	// output "Near cache size is 1"

2. Creating a Near Cache specifying maximum number of entries to store

The following example shows how to get a named cache that will cache up to 100 entries from Get() or GetAll().
When the threshold of HighUnits is reached, the near cache is pruned to the default of 80% of its size and evicts least recently
accessed and created entries.

Note: The default prune percentage is 0.8 (80%) which indicates the percentage of the total number of
units that will remain after the cache manager prunes the near cache( i.e. this is the "low watermark" value).
This can be changed by setting the PruneFactory to a value in the range 0.1 to 1.0 in [NearCacheOptions].

	// specify HighUnits of 1000
	nearCacheOptions := coherence.NearCacheOptions{HighUnits: 1000}

	namedMap, err := coherence.GetNamedMap[int, string](session, "customers", coherence.WithNearCache(&nearCacheOptions))
	if err != nil {
	    log.Fatal(err)
	}

	// assume we have 2000 entries in the coherence cache, issue 1000 gets and the near cache will have 100 entries
	for i := 1; i <= 1000; i++ {
		_, err = namedMap.Get(ctx, i)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Near cache size is", namedMap.GetNearCacheStats().Size())
	// output: "Near cache size is 1000"

	// issue a subsequent Get() for an entry not in the near cache and the cache will be pruned to 80%
	customer, err = namedMap.Get(ctx, 1)

	fmt.Println("Near cache size is", namedCache.GetNearCacheStats().Size())
	// output: "Near cache size is 800"

3. Creating a Near Cache specifying maximum memory to use

The following example shows how to get a named cache that will cache up to 10KB of entries from Get() or GetAll().
When the threshold of HighUnits is reached, the near cache is pruned to 80% of its size and evicts least recently
accessed and created entries.

	// specify HighUnits of 1000
	nearCacheOptions := coherence.NearCacheOptions{HighUnitsMemory: 10 * 1024}

	namedMap, err := coherence.GetNamedMap[int, string](session, "customers", coherence.WithNearCache(&nearCacheOptions))
	if err != nil {
	    log.Fatal(err)
	}

	// assume we have 5000 entries in the coherence cache, issue 5000 gets and the near cache will be pruned and
	// not have the full 5000 entries as it does not fit within 10KB.
	for i := 1; i <= 5000; i++ {
		_, err = namedMap.Get(ctx, i)
		if err != nil {
			panic(err)
		}
	}

	// print the near cache stats via String()
	fmt.Println(namedMap.GetNearCacheStats())
	// localCache{name=my-near-cache-high-units, options=localCacheOptions{ttl=0s, highUnits=1000, highUnitsMemory=0B, pruneFactor=0.80, invalidation=ListenAll}, stats=CacheStats{puts=1001, gets=1002, hits=1, misses=1001, missesDuration=4.628931138s,
	// hitRate=0.0998004, prunes=1, prunesDuration=181.533µs, expires=0, expiresDuration=0s, size=200, memoryUsed=53.2KB}}

[Coherence Documentation]: https://docs.oracle.com/en/middleware/standalone/coherence/14.1.1.2206/develop-applications/introduction-coherence-caches.html
[examples]: https://github.com/oracle/coherence-go-client/tree/main/examples
[gRPC Proxy documentation]: https://docs.oracle.com/en/middleware/standalone/coherence/14.1.1.2206/develop-remote-clients/using-coherence-grpc-server.html
[gRPC Naming]: https://github.com/grpc/grpc/blob/master/doc/naming.md
[Queues]: https://coherence.community/latest/24.09/docs/#/docs/core/09_queues
*/
package coherence
