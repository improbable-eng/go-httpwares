// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`http_retry` is a HTTP client-side Tripperware that allows you to retry requests that are marked as idempotent and safe.

This logic works for requests considered safe and idempotent (configurable) and ones that have no body, or have `Request.GetBody` either automatically implemented (`byte.Buffer`, `string.Buffer`) or specified manually. By default all GET, HEAD and OPTIONS requests are considered idempotent and safe.

The implementation allow retries after client-side errors or returned error codes (by default 5**) according to configurable backoff policies (linear backoff by default). Additionally, requests can be hedged. Hedging works by sending additional requests without waiting for a previous one to return.
*/
package http_retry
