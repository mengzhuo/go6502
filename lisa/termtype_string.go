// Code generated by "stringer --type=TermType --trimprefix=T"; DO NOT EDIT.

package lisa

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TLabel-1]
	_ = x[TOperator-2]
	_ = x[THex-3]
	_ = x[TDecimal-4]
	_ = x[TBinary-5]
	_ = x[TAscii-6]
	_ = x[TCurrentLine-7]
	_ = x[TLSLabel-8]
	_ = x[TGTLabel-9]
	_ = x[TRaw-10]
}

const _TermType_name = "LabelOperatorHexDecimalBinaryAsciiCurrentLineLSLabelGTLabelRaw"

var _TermType_index = [...]uint8{0, 5, 13, 16, 23, 29, 34, 45, 52, 59, 62}

func (i TermType) String() string {
	i -= 1
	if i >= TermType(len(_TermType_index)-1) {
		return "TermType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _TermType_name[_TermType_index[i]:_TermType_index[i+1]]
}
