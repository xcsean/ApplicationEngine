package errno

const (
	// OK means success
	OK = 0

	// SYSINTERNALERROR means some dependencies missed
	SYSINTERNALERROR = -1

	// NODENOTFOUNDINREGISTRY means the node can't be found in registry
	NODENOTFOUNDINREGISTRY = -2

	// NODEIPNOTEQUALREGISTRY means the node ip isn't equal to the ip in registry
	NODEIPNOTEQUALREGISTRY = -3

	// CONNMASTEROFFLINE means no master attached conn
	CONNMASTEROFFLINE = -10

	// CONNMAXCONNECTIONS means connections reach maxlimit
	CONNMAXCONNECTIONS = -11

	// RPCDONOTHAVEPEERINFO means peer info not found in rpc context
	RPCDONOTHAVEPEERINFO = -21

	// HOSTVMALREADYEXIST means the vm identified by division already exist
	HOSTVMALREADYEXIST = -31

	// HOSTVMADDRALREADYEXIST means the vm's address identified by division already exist
	HOSTVMADDRALREADYEXIST = -32

	// HOSTVMUNAVAILABLEBYVER means vm's unavailable by version provided
	HOSTVMUNAVAILABLEBYVER = -33

	// HOSTVMSENDCHANNELFULL means vm's send channel is full
	HOSTVMSENDCHANNELFULL = -34

	// HOSTVMBINDNEEDRETRY means host will notify the vm to bind session again
	HOSTVMBINDNEEDRETRY = -35

	// HOSTVMSESSIONALREADYBIND means the session already was binded
	HOSTVMSESSIONALREADYBIND = -36

	// HOSTVMNOTEXIST means the vm identified by division is not exist in host
	HOSTVMNOTEXIST = -39
)
