_This package was developed by [Silicon Ally](https://siliconally.org) while
working on a project for  [Adventure Scientists](https://adventurescientists.org).
Many thanks to Adventure Scientists for supporting [our open source
mission](https://siliconally.org/policies/open-source/)!_

# gqlerr

`gqlerr` is a package for handling errors in a Go/gqlgen-based GraphQL server.
It integrates logging via *zap.Logger so that all errors are logged. Codes are
modeled after
[the gRPC error codes](https://pkg.go.dev/google.golang.org/grpc/codes).

[![GoDoc](https://pkg.go.dev/badge/github.com/Silicon-Ally/gqlerr?status.svg)](https://pkg.go.dev/github.com/Silicon-Ally/gqlerr?tab=doc)
[![CI Workflow](https://github.com/Silicon-Ally/gqlerr/actions/workflows/test.yml/badge.svg)](https://github.com/Silicon-Ally/gqlerr/actions?query=branch%3Amain)

## Usage

To use this package, set the `gqlerr.ErrorPresenter` as your
`server.SetErrorPresenter` with your already-configured `*zap.Logger` instance.
See [the gqlgen error docs](https://gqlgen.com/reference/errors/) for more
info.

From there, start replacing errors in handlers with calls to `gqlerr`. For
example, if you have code like:

```go
func (r *Resolver) SomeResolver(ctx context.Context, req model.Request) (*model.Response, error) {
  if err := validate(req); err != nil {
    return nil, fmt.Errorf("failed to validate request: %w", err)
  }
  return &model.Response{}, nil
}
```

You'd update it to the following:

```go
func (r *Resolver) SomeResolver(ctx context.Context, req model.Request) (*model.Response, error) {
  if err := validate(req); err != nil {
    return nil, gqlerr.
      InvalidArgument("request failed validation", zap.Error(err)).
      WithMessage("invalid request")
  }
  return &model.Response{}, nil
}
```

This error will get transformed into a standard GraphQL `{"errors": [ ... ]}`,
with the message given by `WithMessage`. By default, a generic message similar
to `net/http`'s `StatusText(...)` function will be returned. A log message will
also be written to the logger, at an appropriate log level for the error code
if one is not explicitly supplied with `At{Debug,Info,Warn,Error}Level()`.
