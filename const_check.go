package gomap

// this is a compile time check for const values. instead a check below
// wich happens every time during an evacuate() func under the hood.
//
//	if evacuatedX+1 != evacuatedY || evacuatedX^1 != evacuatedY {
//		throw("bad evacuatedN")
//	}
func _() {
	var x [1]struct{}
	_ = x[evacuatedFirst-2]
	_ = x[evacuatedSecond-3]
}
