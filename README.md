# proterror

Proto-first error handling for Go services.

`proterror` lets you declare application errors as protobuf messages, annotate
them with canonical gRPC status codes, and generate Go helpers that work with
`errors.As`, gRPC status details, grpc-gateway responses, and service
interceptors.

## Why use it

- Keep error contracts in proto files, next to the APIs that return them.
- Return typed Go errors from services while sending canonical gRPC statuses on
  the wire.
- Decode gRPC status details back into typed errors on clients.
- Hide private/internal errors from clients by falling back to `Unknown`.
- Map common PostgreSQL failures to public API errors.
- Reuse the same error model for gRPC and HTTP gateway responses.

## Packages and tooling

- `github.com/not-for-prod/proterror` provides the runtime helpers,
  interceptors, HTTP writer, and database error mapping.
- `github.com/not-for-prod/proterror/proterror` contains built-in canonical
  error messages such as `NotFound`, `AlreadyExists`, `Internal`, and
  `Unavailable`.
- `github.com/not-for-prod/proterror/registry` stores public generated error
  types used during status conversion.
- `github.com/not-for-prod/proterror/cmd/protoc-gen-proterror` is the protobuf
  generator.

The generator emits `<name>.pb.proterror.go` files for annotated messages.

## Installation

Install the generator with a pinned module version in production builds:

```sh
go install github.com/not-for-prod/proterror/cmd/protoc-gen-proterror@<version>
```

For local development you can use:

```sh
go install github.com/not-for-prod/proterror/cmd/protoc-gen-proterror@latest
```

Add the runtime package to your service:

```sh
go get github.com/not-for-prod/proterror
```

## Buf configuration

Add the generator after the standard Go protobuf and gRPC plugins:

```yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: .
    opt:
      - paths=source_relative
  - remote: buf.build/grpc/go
    out: .
    opt:
      - paths=source_relative
  - local: protoc-gen-proterror
    out: .
    opt:
      - paths=source_relative
```

If you build the plugin inside the repository, point Buf at the binary:

```yaml
  - local: ./bin/protoc-gen-proterror
    out: .
    opt:
      - paths=source_relative
```

## Defining errors

Import `proterror/options.proto` and annotate any protobuf message that should
act as an API error:

```proto
syntax = "proto3";

package example.v1;

import "proterror/options.proto";

message ValidationFailed {
  option (proterror.options) = {
    code: INVALID_ARGUMENT
    internal: false
  };

  string field = 1;
  string reason = 2;
}

message ProviderUnavailable {
  option (proterror.options) = {
    code: UNAVAILABLE
    internal: true
  };
}
```

For each annotated message, the generator adds:

- `Error() string`
- `Is(err error) bool`
- `Code() codes.Code`
- `Internal() bool`
- `Status() *status.Status`
- `Join(err error) error`

Public errors (`internal: false`) are registered automatically and can be
serialized to clients. Internal errors (`internal: true`) are not registered;
when they pass through `AsStatus` or the server interceptor, clients receive the
built-in `Unknown` error instead.

## Server usage

Install the unary server interceptor once when constructing your gRPC server:

```go
package main

import (
	"net"

	proterrors "github.com/not-for-prod/proterror"
	examplev1 "github.com/you/service/gen/example/v1"
	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			proterrors.UnaryServerInterceptor(),
		),
	)

	examplev1.RegisterExampleServiceServer(server, newService())

	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}
```

Return generated errors from handlers:

```go
func (s *Service) Create(ctx context.Context, req *examplev1.CreateRequest) (*examplev1.CreateResponse, error) {
	if req.GetName() == "" {
		return nil, &examplev1.ValidationFailed{
			Field:  "name",
			Reason: "required",
		}
	}

	return &examplev1.CreateResponse{}, nil
}
```

To keep the original cause for logs, tracing, or `errors.Is`/`errors.As`, join
the typed API error with the lower-level error:

```go
func (s *Service) Get(ctx context.Context, req *examplev1.GetRequest) (*examplev1.GetResponse, error) {
	row, err := s.repo.Get(ctx, req.GetId())
	if err != nil {
		return nil, (&examplev1.ValidationFailed{Field: "id"}).Join(err)
	}

	return mapRow(row), nil
}
```

When a joined error reaches the server interceptor, the first registered public
`ProtError` is converted to a gRPC status.

## Client usage

Install the unary client interceptor to decode registered protobuf status
details back into typed errors:

```go
conn, err := grpc.NewClient(
	target,
	grpc.WithChainUnaryInterceptor(
		proterrors.UnaryClientInterceptor(),
	),
)
```

Then handle errors using the standard library:

```go
resp, err := client.Create(ctx, req)
if err != nil {
	var validation *examplev1.ValidationFailed
	if errors.As(err, &validation) {
		return fmt.Errorf("invalid %s: %s", validation.GetField(), validation.GetReason())
	}

	return err
}
```

## HTTP and grpc-gateway

`WriteHTTPResponse` writes a grpc-gateway-compatible JSON response using the
same status code and details model:

```go
func handler(w http.ResponseWriter, r *http.Request) {
	if err := doWork(r.Context()); err != nil {
		proterrors.WriteHTTPResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

Unknown non-`ProtError` values are written as the built-in `Internal` error.

## PostgreSQL mapping

`FromPG` converts common `database/sql` and `pgx` errors into built-in
`proterror` errors while preserving the original error with `errors.Join`:

```go
func (r *Repository) Get(ctx context.Context, id string) (*Entity, error) {
	entity, err := r.query(ctx, id)
	if err != nil {
		return nil, proterrors.FromPG(err)
	}

	return entity, nil
}
```

Current mappings include:

- `sql.ErrNoRows` -> `NotFound`
- unique violations -> `AlreadyExists`
- foreign-key and missing-object failures -> `NotFound`
- invalid input, check, not-null, and range failures -> `InvalidArgument`
- privilege and password failures -> `PermissionDenied` or `Unauthenticated`
- deadlocks and query cancellations -> `DeadlineExceeded`
- connection pressure and unavailable database states -> `Unavailable`
- all other database errors -> `Internal`

## Production guidance

- Pin the generator version in CI and regenerate code as part of your protobuf
  workflow.
- Treat proto error messages as API contracts. Add fields carefully and avoid
  removing or repurposing existing field numbers.
- Mark sensitive failures as `internal: true`; they will not be registered for
  public conversion.
- Put user-safe fields in public errors and keep raw causes in joined errors,
  logs, traces, or metrics.
- Install the server interceptor on every gRPC server entry point that returns
  generated errors.
- Install the client interceptor where callers need typed error handling.
- Test public error behavior at API boundaries, not only in repository or
  service unit tests.

## Development

Repository layout:

- `proterror/` contains the built-in error and option proto definitions.
- `cmd/protoc-gen-proterror/` contains the generator.
- `docs/example/` contains a runnable gRPC example.
- `registry/` contains the runtime public error registry.

Useful commands:

```sh
make generate
go test ./...
```

`make generate` builds `bin/protoc-gen-proterror` and runs `buf generate`.
Install Buf and the protobuf toolchain before regenerating code.
