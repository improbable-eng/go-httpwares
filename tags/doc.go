// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`http_ctxtags` adds a Tag object to the request's context that identifies it for other wares.

Request Context Tags

Tags describe information about the request, and can be set and used by other middleware. Tags are used for logging
and tracing of requests. This extends to both client-side (Tripperware) and server-side (Middleware) libraries.

Service Tags

`http_ctxtags` introduces a concept of services for client-side calls. This makes it easy to identify outbound requests
to both internal and external services. The recommended way to do it is to instantiate a new `http.Client` with a
different `http_ctx.Tripperware(serviceName)` for external services and then reuse that.

Handler Names and Groups

For server-side purposes handlers can be named (e.g. "token_exchange") and placed in a group (e.g. "auth"). This allows
easy organisation of HTTP endpoints for logging and monitoring purposes.


and methods to both server-side and client-side calls. This is to
make it easy to extract semantic meaning of otherwise opaque RESTful URLs, and make it easy to count such requests.

For client-side calls, the service is either an external name (e.g. "github", "aws") or internal name (e.g. "authserver").
A method represents the logical operation (e.g. "place_payment").

For server-side calls, the service is a grouping of http.Handlers, and a method is a logical name for an http.Handler.

See `TagFor*` consts below.

Custom Tags

You can provide a `WithTagExtractor` function that will populate tags server-side and client-side. Each `http.Request`
passing through the Tripperware/Middleware will be passed through this function and new tags will be added.

Tags fields are typed, and shallow and should follow the OpenTracing semantics convention (be prefixed with `http.`):
https://github.com/opentracing/specification/blob/master/semantic_conventions.md
*/

package http_ctxtags
