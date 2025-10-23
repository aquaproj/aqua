---
sidebar_position: 820
---

# Log Color

aqua >= [v1.17.0](https://github.com/aquaproj/aqua/releases/tag/v1.17.0)

[#982](https://github.com/aquaproj/aqua/issues/982) [#983](https://github.com/aquaproj/aqua/pull/983)

aqua supports setting the log coloring by the environment variable `AQUA_LOG_COLOR`.

```console
$ export AQUA_LOG_COLOR=always|auto|never
```

aqua uses [sirupsen/logrus](https://github.com/sirupsen/logrus).
About the color setting, please see [logrus#formatters](https://github.com/sirupsen/logrus#formatters) too.

If you want to disable the coloring, please set `AQUA_LOG_COLOR=never`.

If you want to enable the coloring at GitHub Actions, please set `AQUA_LOG_COLOR=always`.

e.g.

![image](https://user-images.githubusercontent.com/13323303/178093930-6adc8928-96e4-425a-9741-a48aac6ccf75.png)
