package handler

import (
	"fmt"
	"net/http"
)

func serveUnaryPost(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"RPC calls must be made using the HTTP 'POST' method.",
			http.StatusNotImplemented,
		)
	}

	serviceName, methodName, ok := parseMethodPath(r.URL.Path)
	if !ok {
		http.Error(
			w,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
		return
	}

	service, ok := h.services[serviceName]
	if !ok {
		http.Error(
			w,
			fmt.Sprintf("The server does not provide a service named '%s'.", serviceName),
			http.StatusNotFound,
		)
		return
	}

	method, ok := service.MethodByName(methodName)
	if !ok {
		http.Error(
			w,
			fmt.Sprintf("The '%s' service does not have an RPC method named '%s.'", serviceName, methodName),
			http.StatusNotFound,
		)
		return
	}

	if method.InputIsStream() || method.OutputIsStream() {
		http.Error(
			w,
			fmt.Sprintf("The '%s' RPC method uses streaming, which is not supported by this server.", methodName),
			http.StatusMethodNotAllowed,
		)
		return
	}
}
