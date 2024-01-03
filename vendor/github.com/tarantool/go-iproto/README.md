[![Go Reference][godoc-badge]][godoc-url]
[![Actions Status][actions-badge]][actions-url]
[![Telegram][telegram-badge]][telegram-url]
[![Telegram Russian][telegram-badge]][telegramru-url]

# iproto

## Import

```go
import "github.com/tarantool/go-iproto"
```

## Overview

Package `iproto` contains IPROTO constants.

The code generated from Tarantool code. Code generation is only supported for
an actual commit/release. We do not have a goal to support all versions of
Tarantool with a code generator.

## Example

```go
package main

import (
	"fmt"

	"github.com/tarantool/go-iproto"
)

func main() {
	fmt.Printf("%s=%d\n", iproto.ER_READONLY, iproto.ER_READONLY)
	fmt.Printf("%s=%d\n", iproto.IPROTO_FEATURE_WATCHERS, iproto.IPROTO_FEATURE_WATCHERS)
	fmt.Printf("%s=%d\n", iproto.IPROTO_FLAG_COMMIT, iproto.IPROTO_FLAG_COMMIT)
	fmt.Printf("%s=%d\n", iproto.IPROTO_SYNC, iproto.IPROTO_SYNC)
	fmt.Printf("%s=%d\n", iproto.IPROTO_SELECT, iproto.IPROTO_SELECT)
}
```

## Development

You need to install `git` and `go1.13+` first. After that, you need to install
additional dependencies into `$GOBIN`:

```bash
make deps
```

You can generate the code with commands:

```bash
TT_TAG=master make
TT_TAG=3.0.0 make
TT_TAG=master TT_REPO=https://github.com/my/tarantool.git make
```

You need to specify a target branch/tag with the environment variable `TT_TAG`
and you could to specify a repository with the `TT_REPO`.

Makefile has additional targets that can be useful:

```bash
make format
TT_TAG=master make generate
make test
```

A good starting point is [generate.go](./generate.go).

[actions-badge]: https://github.com/tarantool/go-iproto/actions/workflows/test.yml/badge.svg
[actions-url]: https://github.com/tarantool/go-iproto/actions/workflows/test.yml
[godoc-badge]: https://pkg.go.dev/badge/github.com/tarantool/go-iproto.svg
[godoc-url]: https://pkg.go.dev/github.com/tarantool/go-iproto
[telegram-badge]: https://img.shields.io/badge/Telegram-join%20chat-blue.svg
[telegram-url]: http://telegram.me/tarantool
[telegramru-url]: http://telegram.me/tarantoolru
