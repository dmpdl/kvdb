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