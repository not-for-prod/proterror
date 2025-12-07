# proterror

Proto-first error handling for Go services. Annotate your proto messages with a gRPC status code and `protoc-gen-proterror` generates helpers that implement `error`, expose a `codes.Code`, and produce typed `status.Status` instances. Server/client interceptors translate those types into the canonical wire format and fall back to an `Unknown` error when needed.

## What's inside
- Custom proto options in `proterror/options.proto` for attaching `google.rpc.Code` to messages.
- A Buf-powered generator (`protoc-gen-proterror`) that emits `<name>.pb.errors.go` alongside your normal Go stubs.
- Lightweight gRPC unary interceptors that surface `proterror` implementations as gRPC errors.
- A default `Unknown` error message for unexpected failures.

## Getting started
Install the generator locally:
```sh
go install github.com/not-for-prod/proterror/cmd/protoc-gen-proterror@latest
```

Add protogen-plugin into your `buf.gen.yaml`:
```yaml
  - local: protoc-gen-proterror
    out: .
    opt:
      - paths=source_relative
```

## Defining errors in proto
```proto
syntax = "proto3";

package example.v1;

import "google/rpc/code.proto";
import "proterror/options.proto";

message InvalidArgumentError {
  option (proterror.options).code = INVALID_ARGUMENT;
  string field = 1;
}
```

The generator will produce:
- `Error() string` returning the message name
- `Is(err error) bool` for typed comparisons
- `Code() codes.Code` mirroring the declared `google.rpc.Code`
- `Status() *status.Status` with the message attached as details

## Using in a gRPC server
```go
import (
  "context"
  "google.golang.org/grpc"
  "github.com/not-for-prod/proterror"
  pb "github.com/you/yourrepo/gen/example/v1"
)

func main() {
  s := grpc.NewServer(
    grpc.UnaryInterceptor(proterror.UnaryServerInterceptor()),
  )
  // register services...
  _ = s
}

func (svc *Service) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
  return nil, &pb.InvalidArgumentError{Field: "id"} // interceptor converts to gRPC status
}
```

## Client interceptor
`proterror.UnaryClientInterceptor` currently passes calls through unchanged and exists for symmetry; extend it if you need client-side handling of `proterror` types.

## Development
- Proto sources live under `proterror/`.
- `make generate` builds the plugin and runs `buf generate`.
- Requires Buf and protoc tooling available on your PATH.
