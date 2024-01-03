// Code generated by "stringer -type=Flag"; DO NOT EDIT.

package iproto

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[IPROTO_FLAG_COMMIT-1]
	_ = x[IPROTO_FLAG_WAIT_SYNC-2]
	_ = x[IPROTO_FLAG_WAIT_ACK-4]
}

const (
	_Flag_name_0 = "IPROTO_FLAG_COMMITIPROTO_FLAG_WAIT_SYNC"
	_Flag_name_1 = "IPROTO_FLAG_WAIT_ACK"
)

var (
	_Flag_index_0 = [...]uint8{0, 18, 39}
)

func (i Flag) String() string {
	switch {
	case 1 <= i && i <= 2:
		i -= 1
		return _Flag_name_0[_Flag_index_0[i]:_Flag_index_0[i+1]]
	case i == 4:
		return _Flag_name_1
	default:
		return "Flag(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
