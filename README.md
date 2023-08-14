# wslog

#### wslog is a wrapper for slog.

In order to ensure only relying on the standard library, the `lumberjack` package is copied to implement log rolling.

If you don't know how to configure log rolling, please refer to [lumberjack](https://github.com/natefinch/lumberjack).

### Examples

```go
package main

import (
	"github.com/zc2638/wslog"
	"log/slog"
)

func main() {
	cfg := wslog.Confg{
		Format: "json",
		Level:  wslog.LevelInfo,
	}
	l := wslog.New(cfg)
	l.Info("this is a info log")
	l.Log(1, "this is a info+1 log")
}
```

Support digital definition level.  
```go
cfg := wslog.Confg{
    Format: "json",
    Level:  "info+2", // equivalent to `slog.LevelInfo+2`
}
```

Use with context
```go
originLogger := wslog.New(cfg)
ctx := wslog.WithContext(context.Backgroud(), originLogger)
l := wslog.FromContext(ctx)
l.Info("this is a info log")
```

You can get the built-in `level` to realize the `level` change during the running of the program.

```go
level := l.Level().(*slog.LevelVar)
level.Set(slog.LevelInfo)
```

You can use a custom leveler.

```go
level := new(slog.LevelVar)
wslog.New(cfg, wslog.LevelOption(level))
```

You can use a custom handler.

```go
handler := slog.NewJSONHandler(os.Stdout, nil)
wslog.New(cfg, wslog.HandlerOption(handler))
```

You can use a custom writer.

```go
wslog.New(cfg, wslog.WriterOption(os.Stdout))
```

Log rolling is built in by default, you can also use the `lumberjack` package as a `writer`.
```go
w := &lumberjack.Logger{}
wslog.New(cfg, wslog.WriterOption(w))
```

You may want to replace some `key` associated content in the log by default.

```go
replaceAttrFunc := func (groups []string, a slog.Attr) slog.Attr {
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
wslog.New(cfg, wslog.ReplaceAttrOption(replaceAttrFunc))
```

