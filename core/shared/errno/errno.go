package errno

const (
	// OK means success
	OK = 0

	// SYSINTERNALERROR means some dependencies missed
	SYSINTERNALERROR = -1

	// CONNMASTEROFFLINE means no master attached conn
	CONNMASTEROFFLINE = -10

	// CONNMAXCONNECTIONS means connections reach maxlimit
	CONNMAXCONNECTIONS = -11
)