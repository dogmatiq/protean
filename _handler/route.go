package handler

// // parseStandardPath parses the HTTP request path to determine the name of the
// // service and method being requested.
// func parseMethodPath(path string) (service, method string, ok bool) {
// 	// Path must start with a slash.
// 	if len(path) == 0 || path[0] != '/' {
// 		return "", "", false
// 	}

// 	path = path[1:]

// 	// Then we expect another slash at the end of the protocol buffers package
// 	// name.
// 	pkgIndex := strings.IndexByte(path, '/')
// 	if pkgIndex == -1 {
// 		return "", "", false
// 	}

// 	// And another after the service name.
// 	serviceIndex := strings.IndexByte(path[pkgIndex+1:], '/')
// 	if serviceIndex == -1 {
// 		return "", "", false
// 	}

// 	// Anything after the second slash is the method name.
// 	return path[:serviceIndex], path[serviceIndex+1:], true
// }

// // parseMethodPath parses an HTTP request path to determine the name of the
// // service and method being requested.
// func parseMethodPath(path string) (service, method string, ok bool) {
// 	// Path must start with a slash.
// 	if len(path) == 0 || path[0] != '/' {
// 		return "", "", false
// 	}

// 	path = path[1:]

// 	// Then we expect another slash at the end of the protocol buffers package
// 	// name.
// 	pkgIndex := strings.IndexByte(path, '/')
// 	if pkgIndex == -1 {
// 		return "", "", false
// 	}

// 	// And another after the service name.
// 	serviceIndex := strings.IndexByte(path[pkgIndex+1:], '/')
// 	if serviceIndex == -1 {
// 		return "", "", false
// 	}

// 	// Anything after the second slash is the method name.
// 	return path[:serviceIndex], path[serviceIndex+1:], true
// }
