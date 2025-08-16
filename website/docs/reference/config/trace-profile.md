---
sidebar_position: 870
---

# Tracing and CPU Profiling

[#825](https://github.com/aquaproj/aqua/pull/825), aqua >= v1.10.0

You can set global command line options `-trace` and `-cpu-profile` for tracing and CPU Profiling.
This is useful for the performance tuning.

The following Go's standard libraries are used.

* https://pkg.go.dev/runtime/trace
* https://pkg.go.dev/runtime/pprof

## How to use

All sub commands except for `help` and `version` commands support this option.

### Tracing with `runtime/trace`

```console
$ aqua -trace trace.out exec -- gh version # a file trace.out is created
$ go tool trace trace.out
2022/06/01 11:18:47 Parsing trace...
2022/06/01 11:18:47 Splitting trace...
2022/06/01 11:18:47 Opening browser. Trace viewer is listening on http://127.0.0.1:58380
```

![image](https://user-images.githubusercontent.com/13323303/171315748-2ef0945d-ccc0-45f6-af54-b46bdcfb55d6.png)

### CPU Profiling with `runtime/pprof`

```console
$ aqua -cpu-profile pprof.out exec -- gh version # a file pprof.out is created
$ go tool pprof -http=":8000" "$(which aqua)" pprof.out
```

![image](https://user-images.githubusercontent.com/13323303/171329271-c3445a29-6ebc-4740-88fa-2668eeb672f3.png)
