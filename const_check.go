package gomap

func _() {
	// compile time check for changed const values
	var x [1]struct{}
	_ = x[evacuatedFirst-2]
	_ = x[evacuatedSecond-3]
}
