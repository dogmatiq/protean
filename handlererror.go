package protean

import "github.com/dogmatiq/protean/rpcerror"

// unimplementedServiceError returns the RPC error that should be sent to the
// client when it calls a method from an unrecognized service.
func unimplementedServiceError(serviceName string) rpcerror.Error {
	return rpcerror.New(
		rpcerror.NotImplemented,
		"the server does not provide the '%s' service",
		serviceName,
	)
}

// unimplementedServiceError returns the RPC error that should be sent to the
// client when it calls an unrecognized method within a service that is
// recognized.
func unimplementedMethodError(serviceName, methodName string) rpcerror.Error {
	return rpcerror.New(
		rpcerror.NotImplemented,
		"the '%s' service does not contain an RPC method named '%s'",
		serviceName,
		methodName,
	)
}
