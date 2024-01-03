// Package with implementation of methods and structures for work with
// Tarantool instance.
package tarantool

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tarantool/go-iproto"
	"github.com/vmihailenco/msgpack/v5"
)

const requestsMap = 128
const ignoreStreamId = 0
const (
	connDisconnected = 0
	connConnected    = 1
	connShutdown     = 2
	connClosed       = 3
)

const shutdownEventKey = "box.shutdown"

type ConnEventKind int
type ConnLogKind int

var (
	errUnknownRequest = errors.New("the passed connected request doesn't belong " +
		"to the current connection or connection pool")
)

const (
	// Connected signals that connection is established or reestablished.
	Connected ConnEventKind = iota + 1
	// Disconnected signals that connection is broken.
	Disconnected
	// ReconnectFailed signals that attempt to reconnect has failed.
	ReconnectFailed
	// Shutdown signals that shutdown callback is processing.
	Shutdown
	// Either reconnect attempts exhausted, or explicit Close is called.
	Closed

	// LogReconnectFailed is logged when reconnect attempt failed.
	LogReconnectFailed ConnLogKind = iota + 1
	// LogLastReconnectFailed is logged when last reconnect attempt failed,
	// connection will be closed after that.
	LogLastReconnectFailed
	// LogUnexpectedResultId is logged when response with unknown id was received.
	// Most probably it is due to request timeout.
	LogUnexpectedResultId
	// LogWatchEventReadFailed is logged when failed to read a watch event.
	LogWatchEventReadFailed
)

// ConnEvent is sent throw Notify channel specified in Opts.
type ConnEvent struct {
	Conn *Connection
	Kind ConnEventKind
	When time.Time
}

// A raw watch event.
type connWatchEvent struct {
	key   string
	value interface{}
}

var epoch = time.Now()

// Logger is logger type expected to be passed in options.
type Logger interface {
	Report(event ConnLogKind, conn *Connection, v ...interface{})
}

type defaultLogger struct{}

func (d defaultLogger) Report(event ConnLogKind, conn *Connection, v ...interface{}) {
	switch event {
	case LogReconnectFailed:
		reconnects := v[0].(uint)
		err := v[1].(error)
		log.Printf("tarantool: reconnect (%d/%d) to %s failed: %s",
			reconnects, conn.opts.MaxReconnects, conn.Addr(), err)
	case LogLastReconnectFailed:
		err := v[0].(error)
		log.Printf("tarantool: last reconnect to %s failed: %s, giving it up",
			conn.Addr(), err)
	case LogUnexpectedResultId:
		resp := v[0].(*Response)
		log.Printf("tarantool: connection %s got unexpected resultId (%d) in response",
			conn.Addr(), resp.RequestId)
	case LogWatchEventReadFailed:
		err := v[0].(error)
		log.Printf("tarantool: unable to parse watch event: %s", err)
	default:
		args := append([]interface{}{"tarantool: unexpected event ", event, conn}, v...)
		log.Print(args...)
	}
}

// Connection is a handle with a single connection to a Tarantool instance.
//
// It is created and configured with Connect function, and could not be
// reconfigured later.
//
// Connection could be in three possible states:
//
// - In "Connected" state it sends queries to Tarantool.
//
// - In "Disconnected" state it rejects queries with ClientError{Code:
// ErrConnectionNotReady}
//
// - In "Shutdown" state it rejects queries with ClientError{Code:
// ErrConnectionShutdown}. The state indicates that a graceful shutdown
// in progress. The connection waits for all active requests to
// complete.
//
// - In "Closed" state it rejects queries with ClientError{Code:
// ErrConnectionClosed}. Connection could become "Closed" when
// Connection.Close() method called, or when Tarantool disconnected and
// Reconnect pause is not specified or MaxReconnects is specified and
// MaxReconnect reconnect attempts already performed.
//
// You may perform data manipulation operation by calling its methods:
// Call*, Insert*, Replace*, Update*, Upsert*, Call*, Eval*.
//
// In any method that accepts space you my pass either space number or space
// name (in this case it will be looked up in schema). Same is true for index.
//
// ATTENTION: tuple, key, ops and args arguments for any method should be
// and array or should serialize to msgpack array.
//
// ATTENTION: result argument for *Typed methods should deserialize from
// msgpack array, cause Tarantool always returns result as an array.
// For all space related methods and Call16* (but not Call17*) methods Tarantool
// always returns array of array (array of tuples for space related methods).
// For Eval* and Call* Tarantool always returns array, but does not forces
// array of arrays.
//
// If connected to Tarantool 2.10 or newer, connection supports server graceful
// shutdown. In this case, server will wait until all client requests will be
// finished and client disconnects before going down (server also may go down
// by timeout). Client reconnect will happen if connection options enable
// reconnect. Beware that graceful shutdown event initialization is asynchronous.
//
// More on graceful shutdown:
// https://www.tarantool.io/en/doc/latest/dev_guide/internals/iproto/graceful_shutdown/
type Connection struct {
	addr   net.Addr
	dialer Dialer
	c      Conn
	mutex  sync.Mutex
	cond   *sync.Cond
	// schemaResolver contains a SchemaResolver implementation.
	schemaResolver SchemaResolver
	// requestId contains the last request ID for requests with nil context.
	requestId uint32
	// contextRequestId contains the last request ID for requests with context.
	contextRequestId uint32
	// Greeting contains first message sent by Tarantool.
	Greeting *Greeting

	shard      []connShard
	dirtyShard chan uint32

	control chan struct{}
	rlimit  chan struct{}
	opts    Opts
	state   uint32
	dec     *msgpack.Decoder
	lenbuf  [packetLengthBytes]byte

	lastStreamId uint64

	serverProtocolInfo ProtocolInfo
	// watchMap is a map of key -> chan watchState.
	watchMap sync.Map

	// shutdownWatcher is the "box.shutdown" event watcher.
	shutdownWatcher Watcher
	// requestCnt is a counter of active requests.
	requestCnt int64
}

var _ = Connector(&Connection{}) // Check compatibility with connector interface.

type futureList struct {
	first *Future
	last  **Future
}

func (list *futureList) findFuture(reqid uint32, fetch bool) *Future {
	root := &list.first
	for {
		fut := *root
		if fut == nil {
			return nil
		}
		if fut.requestId == reqid {
			if fetch {
				*root = fut.next
				if fut.next == nil {
					list.last = root
				} else {
					fut.next = nil
				}
			}
			return fut
		}
		root = &fut.next
	}
}

func (list *futureList) addFuture(fut *Future) {
	*list.last = fut
	list.last = &fut.next
}

func (list *futureList) clear(err error, conn *Connection) {
	fut := list.first
	list.first = nil
	list.last = &list.first
	for fut != nil {
		fut.SetError(err)
		conn.markDone(fut)
		fut, fut.next = fut.next, nil
	}
}

type connShard struct {
	rmut            sync.Mutex
	requests        [requestsMap]futureList
	requestsWithCtx [requestsMap]futureList
	bufmut          sync.Mutex
	buf             smallWBuf
	enc             *msgpack.Encoder
}

// RLimitActions is an enumeration type for an action to do when a rate limit
// is reached.
type RLimitAction int

const (
	// RLimitDrop immediately aborts the request.
	RLimitDrop RLimitAction = iota
	// RLimitWait waits during timeout period for some request to be answered.
	// If no request answered during timeout period, this request is aborted.
	// If no timeout period is set, it will wait forever.
	RLimitWait
)

// Opts is a way to configure Connection
type Opts struct {
	// Timeout for response to a particular request. The timeout is reset when
	// push messages are received. If Timeout is zero, any request can be
	// blocked infinitely.
	// Also used to setup net.TCPConn.Set(Read|Write)Deadline.
	//
	// Pay attention, when using contexts with request objects,
	// the timeout option for Connection does not affect the lifetime
	// of the request. For those purposes use context.WithTimeout() as
	// the root context.
	Timeout time.Duration
	// Timeout between reconnect attempts. If Reconnect is zero, no
	// reconnect attempts will be made.
	// If specified, then when Tarantool is not reachable or disconnected,
	// new connect attempt is performed after pause.
	// By default, no reconnection attempts are performed,
	// so once disconnected, connection becomes Closed.
	Reconnect time.Duration
	// Maximum number of reconnect failures; after that we give it up to
	// on. If MaxReconnects is zero, the client will try to reconnect
	// endlessly.
	// After MaxReconnects attempts Connection becomes closed.
	MaxReconnects uint
	// RateLimit limits number of 'in-fly' request, i.e. already put into
	// requests queue, but not yet answered by server or timeouted.
	// It is disabled by default.
	// See RLimitAction for possible actions when RateLimit.reached.
	RateLimit uint
	// RLimitAction tells what to do when RateLimit is reached.
	// It is required if RateLimit is specified.
	RLimitAction RLimitAction
	// Concurrency is amount of separate mutexes for request
	// queues and buffers inside of connection.
	// It is rounded up to nearest power of 2.
	// By default it is runtime.GOMAXPROCS(-1) * 4
	Concurrency uint32
	// SkipSchema disables schema loading. Without disabling schema loading,
	// there is no way to create Connection for currently not accessible Tarantool.
	SkipSchema bool
	// Notify is a channel which receives notifications about Connection status
	// changes.
	Notify chan<- ConnEvent
	// Handle is user specified value, that could be retrivied with
	// Handle() method.
	Handle interface{}
	// Logger is user specified logger used for error messages.
	Logger Logger
}

// Connect creates and configures a new Connection.
func Connect(ctx context.Context, dialer Dialer, opts Opts) (conn *Connection, err error) {
	conn = &Connection{
		dialer:           dialer,
		requestId:        0,
		contextRequestId: 1,
		Greeting:         &Greeting{},
		control:          make(chan struct{}),
		opts:             opts,
		dec:              msgpack.NewDecoder(&smallBuf{}),
	}
	maxprocs := uint32(runtime.GOMAXPROCS(-1))
	if conn.opts.Concurrency == 0 || conn.opts.Concurrency > maxprocs*128 {
		conn.opts.Concurrency = maxprocs * 4
	}
	if c := conn.opts.Concurrency; c&(c-1) != 0 {
		for i := uint(1); i < 32; i *= 2 {
			c |= c >> i
		}
		conn.opts.Concurrency = c + 1
	}
	conn.dirtyShard = make(chan uint32, conn.opts.Concurrency*2)
	conn.shard = make([]connShard, conn.opts.Concurrency)
	for i := range conn.shard {
		shard := &conn.shard[i]
		requestsLists := []*[requestsMap]futureList{&shard.requests, &shard.requestsWithCtx}
		for _, requests := range requestsLists {
			for j := range requests {
				requests[j].last = &requests[j].first
			}
		}
	}

	if conn.opts.RateLimit > 0 {
		conn.rlimit = make(chan struct{}, conn.opts.RateLimit)
		if conn.opts.RLimitAction != RLimitDrop && conn.opts.RLimitAction != RLimitWait {
			return nil, errors.New("RLimitAction should be specified to RLimitDone nor RLimitWait")
		}
	}

	if conn.opts.Logger == nil {
		conn.opts.Logger = defaultLogger{}
	}

	conn.cond = sync.NewCond(&conn.mutex)

	if err = conn.createConnection(ctx); err != nil {
		return nil, err
	}

	go conn.pinger()
	if conn.opts.Timeout > 0 {
		go conn.timeouts()
	}

	// TODO: reload schema after reconnect.
	if !conn.opts.SkipSchema {
		schema, err := GetSchema(conn)
		if err != nil {
			conn.mutex.Lock()
			defer conn.mutex.Unlock()
			conn.closeConnection(err, true)
			return nil, err
		}
		conn.SetSchema(schema)
	}

	return conn, err
}

// ConnectedNow reports if connection is established at the moment.
func (conn *Connection) ConnectedNow() bool {
	return atomic.LoadUint32(&conn.state) == connConnected
}

// ClosedNow reports if connection is closed by user or after reconnect.
func (conn *Connection) ClosedNow() bool {
	return atomic.LoadUint32(&conn.state) == connClosed
}

// Close closes Connection.
// After this method called, there is no way to reopen this Connection.
func (conn *Connection) Close() error {
	err := ClientError{ErrConnectionClosed, "connection closed by client"}
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	return conn.closeConnection(err, true)
}

// CloseGraceful closes Connection gracefully. It waits for all requests to
// complete.
// After this method called, there is no way to reopen this Connection.
func (conn *Connection) CloseGraceful() error {
	return conn.shutdown(true)
}

// Addr returns a configured address of Tarantool socket.
func (conn *Connection) Addr() net.Addr {
	return conn.addr
}

// Handle returns a user-specified handle from Opts.
func (conn *Connection) Handle() interface{} {
	return conn.opts.Handle
}

func (conn *Connection) cancelFuture(fut *Future, err error) {
	if fut = conn.fetchFuture(fut.requestId); fut != nil {
		fut.SetError(err)
		conn.markDone(fut)
	}
}

func (conn *Connection) dial(ctx context.Context) error {
	opts := conn.opts

	var c Conn
	c, err := conn.dialer.Dial(ctx, DialOpts{
		IoTimeout: opts.Timeout,
	})
	if err != nil {
		return err
	}

	conn.addr = c.Addr()
	conn.Greeting.Version = c.Greeting().Version
	conn.serverProtocolInfo = c.ProtocolInfo()

	spaceAndIndexNamesSupported :=
		isFeatureInSlice(iproto.IPROTO_FEATURE_SPACE_AND_INDEX_NAMES,
			conn.serverProtocolInfo.Features)

	conn.schemaResolver = &noSchemaResolver{
		SpaceAndIndexNamesSupported: spaceAndIndexNamesSupported,
	}

	// Watchers.
	conn.watchMap.Range(func(key, value interface{}) bool {
		st := value.(chan watchState)
		state := <-st
		if state.unready != nil {
			st <- state
			return true
		}

		req := newWatchRequest(key.(string))
		if err = writeRequest(c, req); err != nil {
			st <- state
			return false
		}
		state.ack = true

		st <- state
		return true
	})

	if err != nil {
		c.Close()
		return fmt.Errorf("unable to register watch: %w", err)
	}

	// Only if connected and fully initialized.
	conn.lockShards()
	conn.c = c
	atomic.StoreUint32(&conn.state, connConnected)
	conn.cond.Broadcast()
	conn.unlockShards()
	go conn.writer(c, c)
	go conn.reader(c, c)

	// Subscribe shutdown event to process graceful shutdown.
	if conn.shutdownWatcher == nil &&
		isFeatureInSlice(iproto.IPROTO_FEATURE_WATCHERS,
			conn.serverProtocolInfo.Features) {
		watcher, werr := conn.newWatcherImpl(shutdownEventKey, shutdownEventCallback)
		if werr != nil {
			return werr
		}
		conn.shutdownWatcher = watcher
	}

	return nil
}

func pack(h *smallWBuf, enc *msgpack.Encoder, reqid uint32,
	req Request, streamId uint64, res SchemaResolver) (err error) {
	const uint32Code = 0xce
	const uint64Code = 0xcf
	const streamBytesLenUint64 = 10
	const streamBytesLenUint32 = 6

	hl := h.Len()

	var streamBytesLen = 0
	var streamBytes [streamBytesLenUint64]byte
	hMapLen := byte(0x82) // 2 element map.
	if streamId != ignoreStreamId {
		hMapLen = byte(0x83) // 3 element map.
		streamBytes[0] = byte(iproto.IPROTO_STREAM_ID)
		if streamId > math.MaxUint32 {
			streamBytesLen = streamBytesLenUint64
			streamBytes[1] = uint64Code
			binary.BigEndian.PutUint64(streamBytes[2:], streamId)
		} else {
			streamBytesLen = streamBytesLenUint32
			streamBytes[1] = uint32Code
			binary.BigEndian.PutUint32(streamBytes[2:], uint32(streamId))
		}
	}

	hBytes := append([]byte{
		uint32Code, 0, 0, 0, 0, // Length.
		hMapLen,
		byte(iproto.IPROTO_REQUEST_TYPE), byte(req.Type()), // Request type.
		byte(iproto.IPROTO_SYNC), uint32Code,
		byte(reqid >> 24), byte(reqid >> 16),
		byte(reqid >> 8), byte(reqid),
	}, streamBytes[:streamBytesLen]...)

	h.Write(hBytes)

	if err = req.Body(res, enc); err != nil {
		return
	}

	l := uint32(h.Len() - 5 - hl)
	h.b[hl+1] = byte(l >> 24)
	h.b[hl+2] = byte(l >> 16)
	h.b[hl+3] = byte(l >> 8)
	h.b[hl+4] = byte(l)

	return
}

func (conn *Connection) createConnection(ctx context.Context) error {
	var err error
	if conn.c == nil && conn.state == connDisconnected {
		if err = conn.dial(ctx); err == nil {
			conn.notify(Connected)
			return nil
		}
	}
	if conn.state == connClosed {
		err = ClientError{ErrConnectionClosed, "using closed connection"}
	}
	return err
}

func (conn *Connection) closeConnection(neterr error, forever bool) (err error) {
	conn.lockShards()
	defer conn.unlockShards()
	if forever {
		if conn.state != connClosed {
			close(conn.control)
			atomic.StoreUint32(&conn.state, connClosed)
			conn.cond.Broadcast()
			// Free the resources.
			if conn.shutdownWatcher != nil {
				go conn.shutdownWatcher.Unregister()
				conn.shutdownWatcher = nil
			}
			conn.notify(Closed)
		}
	} else {
		atomic.StoreUint32(&conn.state, connDisconnected)
		conn.cond.Broadcast()
		conn.notify(Disconnected)
	}
	if conn.c != nil {
		err = conn.c.Close()
		conn.c = nil
	}
	for i := range conn.shard {
		conn.shard[i].buf.Reset()
		requestsLists := []*[requestsMap]futureList{
			&conn.shard[i].requests,
			&conn.shard[i].requestsWithCtx,
		}
		for _, requests := range requestsLists {
			for pos := range requests {
				requests[pos].clear(neterr, conn)
			}
		}
	}
	return
}

func (conn *Connection) getDialTimeout() time.Duration {
	dialTimeout := conn.opts.Reconnect / 2
	if dialTimeout == 0 {
		dialTimeout = 500 * time.Millisecond
	} else if dialTimeout > 5*time.Second {
		dialTimeout = 5 * time.Second
	}
	return dialTimeout
}

func (conn *Connection) runReconnects() error {
	dialTimeout := conn.getDialTimeout()
	var reconnects uint
	var err error

	for conn.opts.MaxReconnects == 0 || reconnects <= conn.opts.MaxReconnects {
		now := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
		err = conn.createConnection(ctx)
		cancel()

		if err != nil {
			if clientErr, ok := err.(ClientError); ok &&
				clientErr.Code == ErrConnectionClosed {
				return err
			}
		} else {
			return nil
		}

		conn.opts.Logger.Report(LogReconnectFailed, conn, reconnects, err)
		conn.notify(ReconnectFailed)
		reconnects++
		conn.mutex.Unlock()

		time.Sleep(time.Until(now.Add(conn.opts.Reconnect)))

		conn.mutex.Lock()
	}

	conn.opts.Logger.Report(LogLastReconnectFailed, conn, err)
	// mark connection as closed to avoid reopening by another goroutine
	return ClientError{ErrConnectionClosed, "last reconnect failed"}
}

func (conn *Connection) reconnectImpl(neterr error, c Conn) {
	if conn.opts.Reconnect > 0 {
		if c == conn.c {
			conn.closeConnection(neterr, false)
			if err := conn.runReconnects(); err != nil {
				conn.closeConnection(err, true)
			}
		}
	} else {
		conn.closeConnection(neterr, true)
	}
}

func (conn *Connection) reconnect(neterr error, c Conn) {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	conn.reconnectImpl(neterr, c)
	conn.cond.Broadcast()
}

func (conn *Connection) lockShards() {
	for i := range conn.shard {
		conn.shard[i].rmut.Lock()
		conn.shard[i].bufmut.Lock()
	}
}

func (conn *Connection) unlockShards() {
	for i := range conn.shard {
		conn.shard[i].rmut.Unlock()
		conn.shard[i].bufmut.Unlock()
	}
}

func (conn *Connection) pinger() {
	to := conn.opts.Timeout
	if to == 0 {
		to = 3 * time.Second
	}
	t := time.NewTicker(to / 3)
	defer t.Stop()
	for {
		select {
		case <-conn.control:
			return
		case <-t.C:
		}
		conn.Ping()
	}
}

func (conn *Connection) notify(kind ConnEventKind) {
	if conn.opts.Notify != nil {
		select {
		case conn.opts.Notify <- ConnEvent{Kind: kind, Conn: conn, When: time.Now()}:
		default:
		}
	}
}

func (conn *Connection) writer(w writeFlusher, c Conn) {
	var shardn uint32
	var packet smallWBuf
	for atomic.LoadUint32(&conn.state) != connClosed {
		select {
		case shardn = <-conn.dirtyShard:
		default:
			runtime.Gosched()
			if len(conn.dirtyShard) == 0 {
				if err := w.Flush(); err != nil {
					err = ClientError{
						ErrIoError,
						fmt.Sprintf("failed to flush data to the connection: %s", err),
					}
					conn.reconnect(err, c)
					return
				}
			}
			select {
			case shardn = <-conn.dirtyShard:
			case <-conn.control:
				return
			}
		}
		shard := &conn.shard[shardn]
		shard.bufmut.Lock()
		if conn.c != c {
			conn.dirtyShard <- shardn
			shard.bufmut.Unlock()
			return
		}
		packet, shard.buf = shard.buf, packet
		shard.bufmut.Unlock()
		if packet.Len() == 0 {
			continue
		}
		if _, err := w.Write(packet.b); err != nil {
			err = ClientError{
				ErrIoError,
				fmt.Sprintf("failed to write data to the connection: %s", err),
			}
			conn.reconnect(err, c)
			return
		}
		packet.Reset()
	}
}

func readWatchEvent(reader io.Reader) (connWatchEvent, error) {
	keyExist := false
	event := connWatchEvent{}
	d := msgpack.NewDecoder(reader)

	l, err := d.DecodeMapLen()
	if err != nil {
		return event, err
	}

	for ; l > 0; l-- {
		cd, err := d.DecodeInt()
		if err != nil {
			return event, err
		}

		switch iproto.Key(cd) {
		case iproto.IPROTO_EVENT_KEY:
			if event.key, err = d.DecodeString(); err != nil {
				return event, err
			}
			keyExist = true
		case iproto.IPROTO_EVENT_DATA:
			if event.value, err = d.DecodeInterface(); err != nil {
				return event, err
			}
		default:
			if err = d.Skip(); err != nil {
				return event, err
			}
		}
	}

	if !keyExist {
		return event, errors.New("watch event does not have a key")
	}

	return event, nil
}

func (conn *Connection) reader(r io.Reader, c Conn) {
	events := make(chan connWatchEvent, 1024)
	defer close(events)

	go conn.eventer(events)

	for atomic.LoadUint32(&conn.state) != connClosed {
		respBytes, err := read(r, conn.lenbuf[:])
		if err != nil {
			err = ClientError{
				ErrIoError,
				fmt.Sprintf("failed to read data from the connection: %s", err),
			}
			conn.reconnect(err, c)
			return
		}
		resp := &Response{buf: smallBuf{b: respBytes}}
		err = resp.decodeHeader(conn.dec)
		if err != nil {
			err = ClientError{
				ErrProtocolError,
				fmt.Sprintf("failed to decode IPROTO header: %s", err),
			}
			conn.reconnect(err, c)
			return
		}

		var fut *Future = nil
		if iproto.Type(resp.Code) == iproto.IPROTO_EVENT {
			if event, err := readWatchEvent(&resp.buf); err == nil {
				events <- event
			} else {
				err = ClientError{
					ErrProtocolError,
					fmt.Sprintf("failed to decode IPROTO_EVENT: %s", err),
				}
				conn.opts.Logger.Report(LogWatchEventReadFailed, conn, err)
			}
			continue
		} else if resp.Code == PushCode {
			if fut = conn.peekFuture(resp.RequestId); fut != nil {
				fut.AppendPush(resp)
			}
		} else {
			if fut = conn.fetchFuture(resp.RequestId); fut != nil {
				fut.SetResponse(resp)
				conn.markDone(fut)
			}
		}

		if fut == nil {
			conn.opts.Logger.Report(LogUnexpectedResultId, conn, resp)
		}
	}
}

// eventer goroutine gets watch events and updates values for watchers.
func (conn *Connection) eventer(events <-chan connWatchEvent) {
	for event := range events {
		if value, ok := conn.watchMap.Load(event.key); ok {
			st := value.(chan watchState)
			state := <-st
			state.value = event.value
			if state.version == math.MaxUint64 {
				state.version = initWatchEventVersion + 1
			} else {
				state.version += 1
			}
			state.ack = false
			if state.changed != nil {
				close(state.changed)
				state.changed = nil
			}
			st <- state
		}
		// It is possible to get IPROTO_EVENT after we already send
		// IPROTO_UNWATCH due to processing on a Tarantool side or slow
		// read from the network, so it looks like an expected behavior.
	}
}

func (conn *Connection) newFuture(ctx context.Context) (fut *Future) {
	fut = NewFuture()
	if conn.rlimit != nil && conn.opts.RLimitAction == RLimitDrop {
		select {
		case conn.rlimit <- struct{}{}:
		default:
			fut.err = ClientError{
				ErrRateLimited,
				"Request is rate limited on client",
			}
			fut.ready = nil
			fut.done = nil
			return
		}
	}
	fut.requestId = conn.nextRequestId(ctx != nil)
	shardn := fut.requestId & (conn.opts.Concurrency - 1)
	shard := &conn.shard[shardn]
	shard.rmut.Lock()
	switch atomic.LoadUint32(&conn.state) {
	case connClosed:
		fut.err = ClientError{
			ErrConnectionClosed,
			"using closed connection",
		}
		fut.ready = nil
		fut.done = nil
		shard.rmut.Unlock()
		return
	case connDisconnected:
		fut.err = ClientError{
			ErrConnectionNotReady,
			"client connection is not ready",
		}
		fut.ready = nil
		fut.done = nil
		shard.rmut.Unlock()
		return
	case connShutdown:
		fut.err = ClientError{
			ErrConnectionShutdown,
			"server shutdown in progress",
		}
		fut.ready = nil
		fut.done = nil
		shard.rmut.Unlock()
		return
	}
	pos := (fut.requestId / conn.opts.Concurrency) & (requestsMap - 1)
	if ctx != nil {
		select {
		case <-ctx.Done():
			fut.SetError(fmt.Errorf("context is done"))
			shard.rmut.Unlock()
			return
		default:
		}
		shard.requestsWithCtx[pos].addFuture(fut)
	} else {
		shard.requests[pos].addFuture(fut)
		if conn.opts.Timeout > 0 {
			fut.timeout = time.Since(epoch) + conn.opts.Timeout
		}
	}
	shard.rmut.Unlock()
	if conn.rlimit != nil && conn.opts.RLimitAction == RLimitWait {
		select {
		case conn.rlimit <- struct{}{}:
		default:
			runtime.Gosched()
			select {
			case conn.rlimit <- struct{}{}:
			case <-fut.done:
				if fut.err == nil {
					panic("fut.done is closed, but err is nil")
				}
			}
		}
	}
	return
}

// This method removes a future from the internal queue if the context
// is "done" before the response is come.
func (conn *Connection) contextWatchdog(fut *Future, ctx context.Context) {
	select {
	case <-fut.done:
	case <-ctx.Done():
	}

	select {
	case <-fut.done:
		return
	default:
		conn.cancelFuture(fut, fmt.Errorf("context is done"))
	}
}

func (conn *Connection) incrementRequestCnt() {
	atomic.AddInt64(&conn.requestCnt, int64(1))
}

func (conn *Connection) decrementRequestCnt() {
	if atomic.AddInt64(&conn.requestCnt, int64(-1)) == 0 {
		conn.cond.Broadcast()
	}
}

func (conn *Connection) send(req Request, streamId uint64) *Future {
	conn.incrementRequestCnt()

	fut := conn.newFuture(req.Ctx())
	if fut.ready == nil {
		conn.decrementRequestCnt()
		return fut
	}

	if req.Ctx() != nil {
		select {
		case <-req.Ctx().Done():
			conn.cancelFuture(fut, fmt.Errorf("context is done"))
			return fut
		default:
		}
		go conn.contextWatchdog(fut, req.Ctx())
	}
	conn.putFuture(fut, req, streamId)

	return fut
}

func (conn *Connection) putFuture(fut *Future, req Request, streamId uint64) {
	shardn := fut.requestId & (conn.opts.Concurrency - 1)
	shard := &conn.shard[shardn]
	shard.bufmut.Lock()
	select {
	case <-fut.done:
		shard.bufmut.Unlock()
		return
	default:
	}
	firstWritten := shard.buf.Len() == 0
	if shard.buf.Cap() == 0 {
		shard.buf.b = make([]byte, 0, 128)
		shard.enc = msgpack.NewEncoder(&shard.buf)
	}
	blen := shard.buf.Len()
	reqid := fut.requestId
	if err := pack(&shard.buf, shard.enc, reqid, req, streamId, conn.schemaResolver); err != nil {
		shard.buf.Trunc(blen)
		shard.bufmut.Unlock()
		if f := conn.fetchFuture(reqid); f == fut {
			fut.SetError(err)
			conn.markDone(fut)
		} else if f != nil {
			/* in theory, it is possible. In practice, you have
			 * to have race condition that lasts hours */
			panic("Unknown future")
		} else {
			fut.wait()
			if fut.err == nil {
				panic("Future removed from queue without error")
			}
			if _, ok := fut.err.(ClientError); ok {
				// packing error is more important than connection
				// error, because it is indication of programmer's
				// mistake.
				fut.SetError(err)
			}
		}
		return
	}
	shard.bufmut.Unlock()

	if firstWritten {
		conn.dirtyShard <- shardn
	}

	if req.Async() {
		if fut = conn.fetchFuture(reqid); fut != nil {
			resp := &Response{
				RequestId: reqid,
				Code:      OkCode,
			}
			fut.SetResponse(resp)
			conn.markDone(fut)
		}
	}
}

func (conn *Connection) markDone(fut *Future) {
	if conn.rlimit != nil {
		<-conn.rlimit
	}
	conn.decrementRequestCnt()
}

func (conn *Connection) peekFuture(reqid uint32) (fut *Future) {
	shard := &conn.shard[reqid&(conn.opts.Concurrency-1)]
	pos := (reqid / conn.opts.Concurrency) & (requestsMap - 1)
	shard.rmut.Lock()
	defer shard.rmut.Unlock()

	if conn.opts.Timeout > 0 {
		if fut = conn.getFutureImp(reqid, true); fut != nil {
			pair := &shard.requests[pos]
			*pair.last = fut
			pair.last = &fut.next
			fut.timeout = time.Since(epoch) + conn.opts.Timeout
		}
	} else {
		fut = conn.getFutureImp(reqid, false)
	}

	return fut
}

func (conn *Connection) fetchFuture(reqid uint32) (fut *Future) {
	shard := &conn.shard[reqid&(conn.opts.Concurrency-1)]
	shard.rmut.Lock()
	fut = conn.getFutureImp(reqid, true)
	shard.rmut.Unlock()
	return fut
}

func (conn *Connection) getFutureImp(reqid uint32, fetch bool) *Future {
	shard := &conn.shard[reqid&(conn.opts.Concurrency-1)]
	pos := (reqid / conn.opts.Concurrency) & (requestsMap - 1)
	// futures with even requests id belong to requests list with nil context
	if reqid%2 == 0 {
		return shard.requests[pos].findFuture(reqid, fetch)
	} else {
		return shard.requestsWithCtx[pos].findFuture(reqid, fetch)
	}
}

func (conn *Connection) timeouts() {
	timeout := conn.opts.Timeout
	t := time.NewTimer(timeout)
	for {
		var nowepoch time.Duration
		select {
		case <-conn.control:
			t.Stop()
			return
		case <-t.C:
		}
		minNext := time.Since(epoch) + timeout
		for i := range conn.shard {
			nowepoch = time.Since(epoch)
			shard := &conn.shard[i]
			for pos := range shard.requests {
				shard.rmut.Lock()
				pair := &shard.requests[pos]
				for pair.first != nil && pair.first.timeout < nowepoch {
					shard.bufmut.Lock()
					fut := pair.first
					pair.first = fut.next
					if fut.next == nil {
						pair.last = &pair.first
					} else {
						fut.next = nil
					}
					fut.SetError(ClientError{
						Code: ErrTimeouted,
						Msg:  fmt.Sprintf("client timeout for request %d", fut.requestId),
					})
					conn.markDone(fut)
					shard.bufmut.Unlock()
				}
				if pair.first != nil && pair.first.timeout < minNext {
					minNext = pair.first.timeout
				}
				shard.rmut.Unlock()
			}
		}
		nowepoch = time.Since(epoch)
		if nowepoch+time.Microsecond < minNext {
			t.Reset(minNext - nowepoch)
		} else {
			t.Reset(time.Microsecond)
		}
	}
}

func read(r io.Reader, lenbuf []byte) (response []byte, err error) {
	var length uint64

	if _, err = io.ReadFull(r, lenbuf); err != nil {
		return
	}
	if lenbuf[0] != 0xce {
		err = errors.New("wrong response header")
		return
	}
	length = (uint64(lenbuf[1]) << 24) +
		(uint64(lenbuf[2]) << 16) +
		(uint64(lenbuf[3]) << 8) +
		uint64(lenbuf[4])

	switch {
	case length == 0:
		err = errors.New("response should not be 0 length")
		return
	case length > math.MaxUint32:
		err = errors.New("response is too big")
		return
	}

	response = make([]byte, length)
	_, err = io.ReadFull(r, response)

	return
}

func (conn *Connection) nextRequestId(context bool) (requestId uint32) {
	if context {
		return atomic.AddUint32(&conn.contextRequestId, 2)
	} else {
		return atomic.AddUint32(&conn.requestId, 2)
	}
}

// Do performs a request asynchronously on the connection.
//
// An error is returned if the request was formed incorrectly, or failed to
// create the future.
func (conn *Connection) Do(req Request) *Future {
	if connectedReq, ok := req.(ConnectedRequest); ok {
		if connectedReq.Conn() != conn {
			fut := NewFuture()
			fut.SetError(errUnknownRequest)
			return fut
		}
	}
	return conn.send(req, ignoreStreamId)
}

// ConfiguredTimeout returns a timeout from connection config.
func (conn *Connection) ConfiguredTimeout() time.Duration {
	return conn.opts.Timeout
}

// SetSchema sets Schema for the connection.
func (conn *Connection) SetSchema(s Schema) {
	sCopy := s.copy()
	spaceAndIndexNamesSupported :=
		isFeatureInSlice(iproto.IPROTO_FEATURE_SPACE_AND_INDEX_NAMES,
			conn.serverProtocolInfo.Features)

	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	conn.lockShards()
	defer conn.unlockShards()

	conn.schemaResolver = &loadedSchemaResolver{
		Schema:                      sCopy,
		SpaceAndIndexNamesSupported: spaceAndIndexNamesSupported,
	}
}

// NewPrepared passes a sql statement to Tarantool for preparation synchronously.
func (conn *Connection) NewPrepared(expr string) (*Prepared, error) {
	req := NewPrepareRequest(expr)
	resp, err := conn.Do(req).Get()
	if err != nil {
		return nil, err
	}
	return NewPreparedFromResponse(conn, resp)
}

// NewStream creates new Stream object for connection.
//
// Since v. 2.10.0, Tarantool supports streams and interactive transactions over them.
// To use interactive transactions, memtx_use_mvcc_engine box option should be set to true.
// Since 1.7.0
func (conn *Connection) NewStream() (*Stream, error) {
	next := atomic.AddUint64(&conn.lastStreamId, 1)
	return &Stream{
		Id:   next,
		Conn: conn,
	}, nil
}

// watchState is the current state of the watcher. See the idea at p. 70, 105:
// https://drive.google.com/file/d/1nPdvhB0PutEJzdCq5ms6UI58dp50fcAN/view
type watchState struct {
	// value is a current value.
	value interface{}
	// version is a current version of the value. The only reason for uint64:
	// go 1.13 has no math.Uint.
	version uint64
	// ack true if the acknowledge is already sent.
	ack bool
	// cnt is a count of active watchers for the key.
	cnt int
	// changed is a channel for broadcast the value changes.
	changed chan struct{}
	// unready channel exists if a state is not ready to work (subscription
	// or unsubscription in progress).
	unready chan struct{}
}

// initWatchEventVersion is an initial version until no events from Tarantool.
const initWatchEventVersion uint64 = 0

// connWatcher is an internal implementation of the Watcher interface.
type connWatcher struct {
	unregister sync.Once
	// done is closed when the watcher is unregistered, but the watcher
	// goroutine is not yet finished.
	done chan struct{}
	// finished is closed when the watcher is unregistered and the watcher
	// goroutine is finished.
	finished chan struct{}
}

// Unregister unregisters the connection watcher.
func (w *connWatcher) Unregister() {
	w.unregister.Do(func() {
		close(w.done)
	})
	<-w.finished
}

// subscribeWatchChannel returns an existing one or a new watch state channel
// for the key. It also increases a counter of active watchers for the channel.
func subscribeWatchChannel(conn *Connection, key string) (chan watchState, error) {
	var st chan watchState

	for st == nil {
		if val, ok := conn.watchMap.Load(key); !ok {
			st = make(chan watchState, 1)
			state := watchState{
				value:   nil,
				version: initWatchEventVersion,
				ack:     false,
				cnt:     0,
				changed: nil,
				unready: make(chan struct{}),
			}
			st <- state

			if val, loaded := conn.watchMap.LoadOrStore(key, st); !loaded {
				if _, err := conn.Do(newWatchRequest(key)).Get(); err != nil {
					conn.watchMap.Delete(key)
					close(state.unready)
					return nil, err
				}
				// It is a successful subsctiption to a watch events by itself.
				state = <-st
				state.cnt = 1
				close(state.unready)
				state.unready = nil
				st <- state
				continue
			} else {
				close(state.unready)
				close(st)
				st = val.(chan watchState)
			}
		} else {
			st = val.(chan watchState)
		}

		// It is an existing channel created outside. It may be in the
		// unready state.
		state := <-st
		if state.unready == nil {
			state.cnt += 1
		}
		st <- state

		if state.unready != nil {
			// Wait for an update and retry.
			<-state.unready
			st = nil
		}
	}

	return st, nil
}

func isFeatureInSlice(expected iproto.Feature, actualSlice []iproto.Feature) bool {
	for _, actual := range actualSlice {
		if expected == actual {
			return true
		}
	}
	return false
}

// NewWatcher creates a new Watcher object for the connection.
//
// Server must support IPROTO_FEATURE_WATCHERS to use watchers.
//
// After watcher creation, the watcher callback is invoked for the first time.
// In this case, the callback is triggered whether or not the key has already
// been broadcast. All subsequent invocations are triggered with
// box.broadcast() called on the remote host. If a watcher is subscribed for a
// key that has not been broadcast yet, the callback is triggered only once,
// after the registration of the watcher.
//
// The watcher callbacks are always invoked in a separate goroutine. A watcher
// callback is never executed in parallel with itself, but they can be executed
// in parallel to other watchers.
//
// If the key is updated while the watcher callback is running, the callback
// will be invoked again with the latest value as soon as it returns.
//
// Watchers survive reconnection. All registered watchers are automatically
// resubscribed when the connection is reestablished.
//
// Keep in mind that garbage collection of a watcher handle doesn’t lead to the
// watcher’s destruction. In this case, the watcher remains registered. You
// need to call Unregister() directly.
//
// Unregister() guarantees that there will be no the watcher's callback calls
// after it, but Unregister() call from the callback leads to a deadlock.
//
// See:
// https://www.tarantool.io/en/doc/latest/reference/reference_lua/box_events/#box-watchers
//
// Since 1.10.0
func (conn *Connection) NewWatcher(key string, callback WatchCallback) (Watcher, error) {
	// We need to check the feature because the IPROTO_WATCH request is
	// asynchronous. We do not expect any response from a Tarantool instance
	// That's why we can't just check the Tarantool response for an unsupported
	// request error.
	if !isFeatureInSlice(iproto.IPROTO_FEATURE_WATCHERS,
		conn.c.ProtocolInfo().Features) {
		err := fmt.Errorf("the feature %s must be supported by connection "+
			"to create a watcher", iproto.IPROTO_FEATURE_WATCHERS)
		return nil, err
	}

	return conn.newWatcherImpl(key, callback)
}

func (conn *Connection) newWatcherImpl(key string, callback WatchCallback) (Watcher, error) {
	st, err := subscribeWatchChannel(conn, key)
	if err != nil {
		return nil, err
	}

	// Start the watcher goroutine.
	done := make(chan struct{})
	finished := make(chan struct{})

	go func() {
		version := initWatchEventVersion
		for {
			state := <-st
			if state.changed == nil {
				state.changed = make(chan struct{})
			}
			st <- state

			if state.version != version {
				callback(WatchEvent{
					Conn:  conn,
					Key:   key,
					Value: state.value,
				})
				version = state.version

				// Do we need to acknowledge the notification?
				state = <-st
				sendAck := !state.ack && version == state.version
				if sendAck {
					state.ack = true
				}
				st <- state

				if sendAck {
					// We expect a reconnect and re-subscribe if it fails to
					// send the watch request. So it looks ok do not check a
					// result. But we need to make sure that the re-watch
					// request will not be finished by a small per-request
					// timeout.
					req := newWatchRequest(key).Context(context.Background())
					conn.Do(req).Get()
				}
			}

			select {
			case <-done:
				state := <-st
				state.cnt -= 1
				if state.cnt == 0 {
					state.unready = make(chan struct{})
				}
				st <- state

				if state.cnt == 0 {
					// The last one sends IPROTO_UNWATCH.
					if !conn.ClosedNow() {
						// conn.ClosedNow() check is a workaround for calling
						// Unregister from connectionClose().
						//
						// We need to make sure that the unwatch request will
						// not be finished by a small per-request timeout to
						// avoid lost of the request.
						req := newUnwatchRequest(key).Context(context.Background())
						conn.Do(req).Get()
					}
					conn.watchMap.Delete(key)
					close(state.unready)
				}

				close(finished)
				return
			case <-state.changed:
			}
		}
	}()

	return &connWatcher{
		done:     done,
		finished: finished,
	}, nil
}

// ProtocolInfo returns protocol version and protocol features
// supported by connected Tarantool server. Beware that values might be
// outdated if connection is in a disconnected state.
// Since 2.0.0
func (conn *Connection) ProtocolInfo() ProtocolInfo {
	return conn.serverProtocolInfo.Clone()
}

func shutdownEventCallback(event WatchEvent) {
	// Receives "true" on server shutdown.
	// See https://www.tarantool.io/en/doc/latest/dev_guide/internals/iproto/graceful_shutdown/
	// step 2.
	val, ok := event.Value.(bool)
	if ok && val {
		go event.Conn.shutdown(false)
	}
}

func (conn *Connection) shutdown(forever bool) error {
	// Forbid state changes.
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	if !atomic.CompareAndSwapUint32(&conn.state, connConnected, connShutdown) {
		if forever {
			err := ClientError{ErrConnectionClosed, "connection closed by client"}
			return conn.closeConnection(err, true)
		}
		return nil
	}

	if forever {
		// We don't want to reconnect any more.
		conn.opts.Reconnect = 0
		conn.opts.MaxReconnects = 0
	}

	conn.cond.Broadcast()
	conn.notify(Shutdown)

	c := conn.c
	for {
		if (atomic.LoadUint32(&conn.state) != connShutdown) || (c != conn.c) {
			return nil
		}
		if atomic.LoadInt64(&conn.requestCnt) == 0 {
			break
		}
		// Use cond var on conn.mutex since request execution may
		// call reconnect(). It is ok if state changes as part of
		// reconnect since Tarantool server won't allow to reconnect
		// in the middle of shutting down.
		conn.cond.Wait()
	}

	if forever {
		err := ClientError{ErrConnectionClosed, "connection closed by client"}
		return conn.closeConnection(err, true)
	} else {
		// Start to reconnect based on common rules, same as in net.box.
		// Reconnect also closes the connection: server waits until all
		// subscribed connections are terminated.
		// See https://www.tarantool.io/en/doc/latest/dev_guide/internals/iproto/graceful_shutdown/
		// step 3.
		conn.reconnectImpl(ClientError{
			ErrConnectionClosed,
			"connection closed after server shutdown",
		}, conn.c)
		return nil
	}
}
