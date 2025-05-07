# KVDB

In memory Key Value DB.

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
### Configuration Structure
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
### Sections

#### Engine Configuration

Controls the storage engine behavior.
| Parameter | Type   | Default     | Description                                                               |
| --------- | ------ | ----------- | ------------------------------------------------------------------------- |
| `type`    | string | "in_memory" | Storage engine type. Currently supported: "in_memory" (RAM-based storage) |

#### Network Configuration
Network-related settings for the server.
| Parameter          | Type     | Default          | Description                                                      |
| ------------------ | -------- | ---------------- | ---------------------------------------------------------------- |
| `address`          | string   | "127.0.0.1:8080" | IP address and port to bind the server to                        |
| `max_connections`  | integer  | 2                | Maximum number of concurrent client connections                  |
| `max_message_size` | string   | "4KB"            | Maximum size of incoming messages (supports KB, MB, GB suffixes) |
| `idle_timeout`     | duration | "5m"             | Connection idle timeout (supports ms, s, m, h suffixes)          |


####  Logging Configuration
Controls logging behavior.

### WAL Configuration (Optional)
Write-Ahead Log settings for persistence. If not specified, WAL will be disabled.

| Parameter                | Type     | Default  | Description                                                      |
| ------------------------ | -------- | -------- | ---------------------------------------------------------------- |
| `flushing_batch_length`  | integer  | 100      | Number of operations to batch before flushing to disk            |
| `flushing_batch_timeout` | duration | "10ms"   | Maximum time to wait before flushing a partial batch             |
| `max_segment_size`       | string   | "1KB"    | Maximum size of WAL segment files (supports KB, MB, GB suffixes) |
| `data_directory`         | string   | "./.wal" | Directory to store WAL segment files                             |

### Duration Format
Time durations can be specified as:
```
Milliseconds: "100ms"
Seconds: "30s"
Minutes: "5m"
Hours: "2h"
```

### Size Format
Data sizes can be specified as:
```
Kilobytes: "4KB"
Megabytes: "2MB"
Gigabytes: "1GB"
```

### Example Configurations
Minimal Configuration
```yaml
engine:
  type: "in_memory"
```

Production-like Configuration
```yaml
engine:
  type: "in_memory"
  
network:
  address: "0.0.0.0:8080"
  max_connections: 1000
  max_message_size: "1MB"
  idle_timeout: "10m"

logging:
  level: "warn"
  output: "/var/log/kvstore.log"

wal:
  flushing_batch_length: 500
  flushing_batch_timeout: "100ms"
  max_segment_size: "10MB"
  data_directory: "/data/wal"
```

## How to run
`make all` - run test, lint code and run server with default config placed in `etc/server.yaml`.

`make run-client` - start database client.
