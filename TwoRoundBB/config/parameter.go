package config

const (
	F   int = 1
	N       = 4
	QC0     = 4*F - 1
	QC1     = 4*F - 1
	//QC2     = 4*F - 1 // prior size
	QC2 = 2*F + 1 // new size

	Size = 250 // the byte length of a value
	T    = 5   // simulate network delay, in ms

	D int64 = 2 // trigger timeout \Delta in second
)
