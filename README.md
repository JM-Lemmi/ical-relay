ical-relay
==========
Relay ical event url and exclude events based on a regex.

# Usage
* Run from source: `go run .`
* Build and run: `go build . && ./ical-relay`

# Config
```toml
url = "https://example.com/events.ical"

[profiles]
    [profiles.p1]
    regex = ["pattern1", "pattern2"]
    public = true
```
Access filtered stream on `/profiles/p1`
