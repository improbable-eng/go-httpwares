// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`http_ctxtags` adds a Tag object to the request's context that identifies it for other wares.

Request Context Tags

Tags describe information about the request, and can be set and used by other middleware. Tags are used for logging
and tracing of requests. This extends to both client-side (Tripperware) and server-side (Middleware) libraries.

Service and Method Tags

`http_ctxtags` introduces a concept of services and methods to both server-side and client-side calls. This is to
make it easy to extract semantic meaning of otherwise opaque RESTful URLs, and make it easy to count such requests.

For client-side calls, the service is either an external name (e.g. "github", "aws") or internal name (e.g. "authserver").
A method represents the logical operation (e.g. "place_payment").

For server-side calls, the service is a grouping of http.Handlers, and a method is a logical name for an http.Handler.

See `TagFor*` consts below.

Custom Tags

You can provide a `WithTagExtractor` function that will populate tags server-side and client-side. These will be exposed
to all the monitoring and logging middleware, and added to traces by default.

Tags fields are typed, and shallow and should follow the OpenTracing semantics convention:
https://github.com/opentracing/specification/blob/master/semantic_conventions.md
*/

package http_ctxtags
