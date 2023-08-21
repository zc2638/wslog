# wslog

#### wslog is a wrapper for slog.

In order to ensure only relying on the standard library, the `lumberjack` package is copied to implement log rolling.

If you don't know how to configure log rolling, please refer to [lumberjack](https://github.com/natefinch/lumberjack).

### Installation

```shell
go get -u github.com/zc2638/wslog
```

### Definition

#### Format

- `json` represents the JSON format log
- `text` represents the Text format log
- others represent the default Log format log

### Examples

```go
package main

import (
	"github.com/zc2638/wslog"
)

func main() {
	cfg := wslog.Confg{
		Format: "json",
		Level:  "info",
	}
	l := wslog.New(cfg)
	l.Info("the info log")
	l.Log(wslog.LevelInfo+1, "the info+1 log")
	l.Log(1, "another info+1 log")
}
```

Support digital definition level.

```go
cfg := wslog.Confg{
    Format: "json",
    Level:  "info+2", // equivalent to `wslog.LevelInfo+2`
}
```

Use with context

```go
logger := wslog.New(cfg)
ctx := wslog.WithContext(context.Backgroud(), logger)
l := wslog.FromContext(ctx)
l.Info("the info log")
```

You can get the built-in `level` to realize the `level` change during the running of the program.

```go
level := l.Level().(*wslog.LevelVar)
level.Set(LevelInfo)
```

You can use a custom `Leveler`.

```go
level := new(wslog.LevelVar)
wslog.New(cfg, level)
```

You can use a custom `HandlerOptions`.

```go
cfg := wslog.Confg{
    Format: "text",
    Level:  "info",
}
handlerOptions := cfg.HandlerOptions()
wslog.New(cfg, handlerOptions)
```

You can use a custom `Handler`.

```go
handler := wslog.NewLogHandler(os.Stdout, nil)
wslog.New(cfg, handler)
```

You can use a custom `io.Writer`.

```go
wslog.New(cfg, io.Writer(os.Stdout))
```

Log rolling is built in by default, you can also use the `lumberjack` package as a `writer`.

```go
w := &lumberjack.Logger{}
wslog.New(cfg, io.Writer(w))
```

You may want to replace some `key` associated content in the log by default.

```go
replaceAttrFunc := func (groups []string, a Attr) Attr {
    // Remove time.
    if a.Key == slog.TimeKey && len(groups) == 0 {
        return slog.Attr{}
    }
    // Remove the directory from the source's filename.
    if a.Key == slog.SourceKey {
        source := a.Value.Any().(*slog.Source)
        source.File = filepath.Base(source.File)
    }
    return a
}
wslog.New(cfg, replaceAttrFunc)
```

You can combine multiple output sources.

```go
h1 := wslog.NewLogHandler(os.Stdout, nil, false)
h2 := slog.NewJSONHandler(os.Stdout, nil)
h3 := NewKafkaHandler() // custom writer, like kafka
multiHandler := wslog.NewMultiHandler(h1, h2, h3)
wslog.New(cfg, multiHandler)
```
