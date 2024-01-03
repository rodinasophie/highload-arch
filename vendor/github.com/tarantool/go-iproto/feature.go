// Code generated by generate.sh; DO NOT EDIT.

package iproto

// IPROTO feature constants, generated from
// tarantool/src/box/iproto_features.h
type Feature int

const (
	// Streams support: IPROTO_STREAM_ID header key.
	IPROTO_FEATURE_STREAMS Feature = 0
	// Transactions in the protocol:
	// IPROTO_BEGIN, IPROTO_COMMIT, IPROTO_ROLLBACK commands.
	IPROTO_FEATURE_TRANSACTIONS Feature = 1
	// MP_ERROR MsgPack extension support.
	//
	// If a client doesn't set this feature bit, then errors returned by
	// CALL/EVAL commands will be encoded according to the serialization
	// rules for generic cdata/userdata Lua objects irrespective of the
	// value of the msgpack.cfg.encode_errors_as_ext flag (by default
	// converted to a string error message). If the feature bit is set and
	// encode_errors_as_ext is true, errors will be encoded as MP_ERROR
	// MsgPack extension.
	IPROTO_FEATURE_ERROR_EXTENSION Feature = 2
	// Remote watchers support:
	// IPROTO_WATCH, IPROTO_UNWATCH, IPROTO_EVENT commands.
	IPROTO_FEATURE_WATCHERS Feature = 3
	// Pagination support:
	// IPROTO_AFTER_POSITION, IPROTO_AFTER_TUPLE, IPROTO_FETCH_POSITION
	// request fields and IPROTO_POSITION response field.
	IPROTO_FEATURE_PAGINATION Feature = 4
	// Using space [index] names instead of identifiers support:
	// IPROTO_SPACE_NAME and IPROTO_INDEX_NAME fields in IPROTO_SELECT,
	// IPROTO_UPDATE and IPROTO_DELETE request body;
	// IPROTO_SPACE_NAME field in IPROTO_INSERT, IPROTO_REPLACE,
	// IPROTO_UPDATE and IPROTO_UPSERT request body.
	IPROTO_FEATURE_SPACE_AND_INDEX_NAMES Feature = 5
	// IPROTO_WATCH_ONCE request support.
	IPROTO_FEATURE_WATCH_ONCE Feature = 6
	// Tuple format in DML request responses support:
	// Tuples in IPROTO_DATA response field are encoded as MP_TUPLE and
	// tuple format is sent in IPROTO_TUPLE_FORMATS field.
	IPROTO_FEATURE_DML_TUPLE_EXTENSION Feature = 7
	// Tuple format in call and eval request responses support:
	// Tuples in IPROTO_DATA response field are encoded as MP_TUPLE and
	// tuple formats are sent in IPROTO_TUPLE_FORMATS field.
	IPROTO_FEATURE_CALL_RET_TUPLE_EXTENSION Feature = 8
	// Tuple format in call and eval request arguments support:
	// Tuples in IPROTO_TUPLE request field are encoded as MP_TUPLE and
	// tuple formats are received in IPROTO_TUPLE_FORMATS field.
	IPROTO_FEATURE_CALL_ARG_TUPLE_EXTENSION Feature = 9
)