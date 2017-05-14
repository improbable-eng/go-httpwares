// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`http_ctxtags` adds a Tag object to the request's context that identifies it for other wares.

Request Context Tags

Tags describe information about the request, and can be set and used by other middleware. Tags are used for logging
and tracing of requests. This extends to both client-side (Tripperware) and server-side (Middleware) libraries.

Service Tags

`http_ctxtags` introduces a concept of services for client-side calls. This makes it easy to identify outbound requests
to both internal and external services.

For calling external services a `http_ctx.Tripperware()` will automatically try to guess the service name from the URL,
e.g. calling "www.googleapis.com" will yield `googleapis` as service name.

However, for calling internal services, it is recommended to explicitly state the service name by using `WithServiceName("myservice")`
and reusing that particular client for all subsequent calls.

Handler Names and Groups

For server-side purposes handlers can be named (e.g. "token_exchange") and placed in a group (e.g. "auth"). This allows
easy organisation of HTTP endpoints for logging and monitoring purposes.

See `TagFor*` consts below.

Custom Tags

You can provide a `WithTagExtractor` function that will populate tags server-side and client-side. Each `http.Request`
passing through the Tripperware/Middleware will be passed through this function and new tags will be added.

Tags fields are typed, and shallow and should follow the OpenTracing semantics convention (be prefixed with `http.`):
https://github.com/opentracing/specification/blob/master/semantic_conventions.md
*/
package http_ctxtags
