package tarantool

import (
	"fmt"

	"github.com/tarantool/go-iproto"
	"github.com/vmihailenco/msgpack/v5"
)

type Response struct {
	RequestId uint32
	Code      uint32
	// Error contains an error message.
	Error string
	// Data contains deserialized data for untyped requests.
	Data []interface{}
	// Pos contains a position descriptor of last selected tuple.
	Pos      []byte
	MetaData []ColumnMetaData
	SQLInfo  SQLInfo
	buf      smallBuf
}

type ColumnMetaData struct {
	FieldName            string
	FieldType            string
	FieldCollation       string
	FieldIsNullable      bool
	FieldIsAutoincrement bool
	FieldSpan            string
}

type SQLInfo struct {
	AffectedCount        uint64
	InfoAutoincrementIds []uint64
}

func (meta *ColumnMetaData) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeMapLen(); err != nil {
		return err
	}
	if l == 0 {
		return fmt.Errorf("map len doesn't match: %d", l)
	}
	for i := 0; i < l; i++ {
		var mk uint64
		var mv interface{}
		if mk, err = d.DecodeUint64(); err != nil {
			return fmt.Errorf("failed to decode meta data")
		}
		if mv, err = d.DecodeInterface(); err != nil {
			return fmt.Errorf("failed to decode meta data")
		}
		switch iproto.MetadataKey(mk) {
		case iproto.IPROTO_FIELD_NAME:
			meta.FieldName = mv.(string)
		case iproto.IPROTO_FIELD_TYPE:
			meta.FieldType = mv.(string)
		case iproto.IPROTO_FIELD_COLL:
			meta.FieldCollation = mv.(string)
		case iproto.IPROTO_FIELD_IS_NULLABLE:
			meta.FieldIsNullable = mv.(bool)
		case iproto.IPROTO_FIELD_IS_AUTOINCREMENT:
			meta.FieldIsAutoincrement = mv.(bool)
		case iproto.IPROTO_FIELD_SPAN:
			meta.FieldSpan = mv.(string)
		default:
			return fmt.Errorf("failed to decode meta data")
		}
	}
	return nil
}

func (info *SQLInfo) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeMapLen(); err != nil {
		return err
	}
	if l == 0 {
		return fmt.Errorf("map len doesn't match")
	}
	for i := 0; i < l; i++ {
		var mk uint64
		if mk, err = d.DecodeUint64(); err != nil {
			return fmt.Errorf("failed to decode meta data")
		}
		switch iproto.SqlInfoKey(mk) {
		case iproto.SQL_INFO_ROW_COUNT:
			if info.AffectedCount, err = d.DecodeUint64(); err != nil {
				return fmt.Errorf("failed to decode meta data")
			}
		case iproto.SQL_INFO_AUTOINCREMENT_IDS:
			if err = d.Decode(&info.InfoAutoincrementIds); err != nil {
				return fmt.Errorf("failed to decode meta data")
			}
		default:
			return fmt.Errorf("failed to decode meta data")
		}
	}
	return nil
}

func (resp *Response) smallInt(d *msgpack.Decoder) (i int, err error) {
	b, err := resp.buf.ReadByte()
	if err != nil {
		return
	}
	if b <= 127 {
		return int(b), nil
	}
	resp.buf.UnreadByte()
	return d.DecodeInt()
}

func (resp *Response) decodeHeader(d *msgpack.Decoder) (err error) {
	var l int
	d.Reset(&resp.buf)
	if l, err = d.DecodeMapLen(); err != nil {
		return
	}
	for ; l > 0; l-- {
		var cd int
		if cd, err = resp.smallInt(d); err != nil {
			return
		}
		switch iproto.Key(cd) {
		case iproto.IPROTO_SYNC:
			var rid uint64
			if rid, err = d.DecodeUint64(); err != nil {
				return
			}
			resp.RequestId = uint32(rid)
		case iproto.IPROTO_REQUEST_TYPE:
			var rcode uint64
			if rcode, err = d.DecodeUint64(); err != nil {
				return
			}
			resp.Code = uint32(rcode)
		default:
			if err = d.Skip(); err != nil {
				return
			}
		}
	}
	return nil
}

func (resp *Response) decodeBody() (err error) {
	if resp.buf.Len() > 2 {
		offset := resp.buf.Offset()
		defer resp.buf.Seek(offset)

		var l, larr int
		var stmtID, bindCount uint64
		var serverProtocolInfo ProtocolInfo
		var feature iproto.Feature
		var errorExtendedInfo *BoxError = nil

		d := msgpack.NewDecoder(&resp.buf)
		d.SetMapDecoder(func(dec *msgpack.Decoder) (interface{}, error) {
			return dec.DecodeUntypedMap()
		})

		if l, err = d.DecodeMapLen(); err != nil {
			return err
		}
		for ; l > 0; l-- {
			var cd int
			if cd, err = resp.smallInt(d); err != nil {
				return err
			}
			switch iproto.Key(cd) {
			case iproto.IPROTO_DATA:
				var res interface{}
				var ok bool
				if res, err = d.DecodeInterface(); err != nil {
					return err
				}
				if resp.Data, ok = res.([]interface{}); !ok {
					return fmt.Errorf("result is not array: %v", res)
				}
			case iproto.IPROTO_ERROR:
				if errorExtendedInfo, err = decodeBoxError(d); err != nil {
					return err
				}
			case iproto.IPROTO_ERROR_24:
				if resp.Error, err = d.DecodeString(); err != nil {
					return err
				}
			case iproto.IPROTO_SQL_INFO:
				if err = d.Decode(&resp.SQLInfo); err != nil {
					return err
				}
			case iproto.IPROTO_METADATA:
				if err = d.Decode(&resp.MetaData); err != nil {
					return err
				}
			case iproto.IPROTO_STMT_ID:
				if stmtID, err = d.DecodeUint64(); err != nil {
					return err
				}
			case iproto.IPROTO_BIND_COUNT:
				if bindCount, err = d.DecodeUint64(); err != nil {
					return err
				}
			case iproto.IPROTO_VERSION:
				if err = d.Decode(&serverProtocolInfo.Version); err != nil {
					return err
				}
			case iproto.IPROTO_FEATURES:
				if larr, err = d.DecodeArrayLen(); err != nil {
					return err
				}

				serverProtocolInfo.Features = make([]iproto.Feature, larr)
				for i := 0; i < larr; i++ {
					if err = d.Decode(&feature); err != nil {
						return err
					}
					serverProtocolInfo.Features[i] = feature
				}
			case iproto.IPROTO_AUTH_TYPE:
				var auth string
				if auth, err = d.DecodeString(); err != nil {
					return err
				}
				found := false
				for _, a := range [...]Auth{ChapSha1Auth, PapSha256Auth} {
					if auth == a.String() {
						serverProtocolInfo.Auth = a
						found = true
					}
				}
				if !found {
					return fmt.Errorf("unknown auth type %s", auth)
				}
			case iproto.IPROTO_POSITION:
				if resp.Pos, err = d.DecodeBytes(); err != nil {
					return fmt.Errorf("unable to decode a position: %w", err)
				}
			default:
				if err = d.Skip(); err != nil {
					return err
				}
			}
		}
		if stmtID != 0 {
			stmt := &Prepared{
				StatementID: PreparedID(stmtID),
				ParamCount:  bindCount,
				MetaData:    resp.MetaData,
			}
			resp.Data = []interface{}{stmt}
		}

		// Tarantool may send only version >= 1
		if serverProtocolInfo.Version != ProtocolVersion(0) || serverProtocolInfo.Features != nil {
			if serverProtocolInfo.Version == ProtocolVersion(0) {
				return fmt.Errorf("no protocol version provided in Id response")
			}
			if serverProtocolInfo.Features == nil {
				return fmt.Errorf("no features provided in Id response")
			}
			resp.Data = []interface{}{serverProtocolInfo}
		}

		if resp.Code != OkCode && resp.Code != PushCode {
			resp.Code &^= uint32(iproto.IPROTO_TYPE_ERROR)
			err = Error{iproto.Error(resp.Code), resp.Error, errorExtendedInfo}
		}
	}
	return
}

func (resp *Response) decodeBodyTyped(res interface{}) (err error) {
	if resp.buf.Len() > 0 {
		offset := resp.buf.Offset()
		defer resp.buf.Seek(offset)

		var errorExtendedInfo *BoxError = nil

		var l int

		d := msgpack.NewDecoder(&resp.buf)
		d.SetMapDecoder(func(dec *msgpack.Decoder) (interface{}, error) {
			return dec.DecodeUntypedMap()
		})

		if l, err = d.DecodeMapLen(); err != nil {
			return err
		}
		for ; l > 0; l-- {
			var cd int
			if cd, err = resp.smallInt(d); err != nil {
				return err
			}
			switch iproto.Key(cd) {
			case iproto.IPROTO_DATA:
				if err = d.Decode(res); err != nil {
					return err
				}
			case iproto.IPROTO_ERROR:
				if errorExtendedInfo, err = decodeBoxError(d); err != nil {
					return err
				}
			case iproto.IPROTO_ERROR_24:
				if resp.Error, err = d.DecodeString(); err != nil {
					return err
				}
			case iproto.IPROTO_SQL_INFO:
				if err = d.Decode(&resp.SQLInfo); err != nil {
					return err
				}
			case iproto.IPROTO_METADATA:
				if err = d.Decode(&resp.MetaData); err != nil {
					return err
				}
			case iproto.IPROTO_POSITION:
				if resp.Pos, err = d.DecodeBytes(); err != nil {
					return fmt.Errorf("unable to decode a position: %w", err)
				}
			default:
				if err = d.Skip(); err != nil {
					return err
				}
			}
		}
		if resp.Code != OkCode && resp.Code != PushCode {
			resp.Code &^= uint32(iproto.IPROTO_TYPE_ERROR)
			err = Error{iproto.Error(resp.Code), resp.Error, errorExtendedInfo}
		}
	}
	return
}

// String implements Stringer interface.
func (resp *Response) String() (str string) {
	if resp.Code == OkCode {
		return fmt.Sprintf("<%d OK %v>", resp.RequestId, resp.Data)
	}
	return fmt.Sprintf("<%d ERR 0x%x %s>", resp.RequestId, resp.Code, resp.Error)
}

// Tuples converts result of Eval and Call to same format
// as other actions returns (i.e. array of arrays).
func (resp *Response) Tuples() (res [][]interface{}) {
	res = make([][]interface{}, len(resp.Data))
	for i, t := range resp.Data {
		switch t := t.(type) {
		case []interface{}:
			res[i] = t
		default:
			res[i] = []interface{}{t}
		}
	}
	return res
}
