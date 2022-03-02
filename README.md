# gqlerr

`gqlerr` is a package for, well, handling GraphQL errors, in the same vein as
our [Silicon-Ally/grpcerr](https://github.com/Silicon-Ally/grpcerr) repository.

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
