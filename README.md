# KVDB

## Command
```
query = set_command | get_command | del_command

set_command = "SET" argument argument
get_command = "GET" argument
del_command = "DEL" argument
argument    = punctuation | letter | digit { punctuation | letter | digit }

punctuation = "\*" | "/" | "_" | ...
letter      = "a" | ... | "z" | "A" | ... | "Z"
digit       = "0" | ... | "9"
```

examples:
```
SET weather_2_pm cold_moscow_weather
GET /etc/nginx/config
DEL user_\*\*\*\*
```

## Configuration

```yaml
engine:
  type: "in_memory"
network:
  address: "127.0.0.1:3223"
  max_connections: 100
  max_message_size: "4KB"
  idle_timeout: 5m
logging:
  level: "info"
  output: "/log/output.log"
```

## How to run
`make all` - run test, lint code and run server with default config placed in `etc/server.yaml`.
