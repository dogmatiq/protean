package handler

// import (
// 	"net/http"

// 	"github.com/dogmatiq/harpy/codegenapi"
// )

// const (
// 	nativeContentType  = "application/harpy"
// 	jsonRPCContentType = "application/json-rpc"
// )

// // GetHandler is an HTTP handler that dispatches GET requests to a service.
// //
// // The RPC method name is indicated by the HTTP request URL's path component.
// //
// // The RPC request message is produced by inspecting HTTP query parameters.
// // [TODO: specify and document the parameter -> message mapping].
// //
// // RPC methods that use client streaming (that is, they accept a stream of
// // requests) are supplied a single request. If no query parameters are provided
// // no request is supplied.
// //
// // Regular RPC methods that do not use client streaming are passed a nil request
// // when there are no query parameters.
// type GetHandler struct {
// 	Service codegenapi.Service
// }

// func (h *GetHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
// 	if req.Method != http.MethodGet {
// 		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	m, ok := h.Service.LookupMethod(req.URL.Path)
// 	if !ok {
// 		http.Error(w, "method not found", http.StatusNotFound)
// 		return
// 	}
// }
