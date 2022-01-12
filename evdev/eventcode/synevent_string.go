// Code generated by "stringer -type=SynEvent"; DO NOT EDIT.

package eventcode

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[SYN_REPORT-0]
	_ = x[SYN_CONFIG-1]
	_ = x[SYN_MT_REPORT-2]
	_ = x[SYN_DROPPED-3]
	_ = x[SYN_MAX-15]
	_ = x[SYN_CNT-16]
}

const (
	_SynEvent_name_0 = "SYN_REPORTSYN_CONFIGSYN_MT_REPORTSYN_DROPPED"
	_SynEvent_name_1 = "SYN_MAXSYN_CNT"
)

var (
	_SynEvent_index_0 = [...]uint8{0, 10, 20, 33, 44}
	_SynEvent_index_1 = [...]uint8{0, 7, 14}
)

func (i SynEvent) String() string {
	switch {
	case 0 <= i && i <= 3:
		return _SynEvent_name_0[_SynEvent_index_0[i]:_SynEvent_index_0[i+1]]
	case 15 <= i && i <= 16:
		i -= 15
		return _SynEvent_name_1[_SynEvent_index_1[i]:_SynEvent_index_1[i+1]]
	default:
		return "SynEvent(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}