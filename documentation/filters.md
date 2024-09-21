### Filters

Currently the Filters can not handle Repeating Events. See Issue #77

#### regex

* `regex`: The regex to match against
* `target`: Parameter to match against the regex. Default Summary, options: Summary, Description, Location

#### id

* `id`: Event ID.

This can match multiple events, for example with repeating events.

#### timeframe

* `after`, `before`. At least one is mandatory. Uses max time, if none is given. Can also be set to "now".

#### duplicates

No parameters. Filters the second and following events that are identified as duplicate. Looks at start, end, summary. If all three are equal, the Event is deemed duplicate.

#### all

No parameters. Filters all.

#### duration

* `duration` in timeDuration format (most relevant: `m`, `h`)´
* `operator`. Either "longer" or "shorter", default "longer".

### Actions

#### delete

No parameters. Deletes the Filtered Events.

#### edit

* `new-summary`, optional: the new summary
* `new-description`, optional: the new description
* `new-start`, optional: the new start time in RFC3339 format "2006-01-02T15:04:05Z"
* `new-end`, optional: the new end time in RFC3339 format "2006-01-02T15:04:05Z"
* `new-location`, optional: the new location
* `overwrite`, default true: Possible values are 'true', 'false' and 'fillempty'. True: Overwrite the property if it already exists; False: Append, Fillempty: Only fills empty properties.  Does not apply to 'new-start' and 'new-end'.
* `move-time`, optional, not together with 'new-start' or 'new-end': add time to the whole entry, to move entry. uses Go ParseDuration: most useful units are "m", "h"
  * when the original time does not have a timezone, sets the timezone to UTC, so it needs to be adjusted for that.

#### add-reminder

* `time`: time in timeDuration Format the alarm will go off before the event.

This usually doesnt work when used in server mode. Most Calendar Applications ignore reminders of external calendars.

#### strip-info

* `mode`: "availibility" (puts busy status as summary, and removes all other information), or "limited" (only keeps summary and busy status)

Inspired by Outlooks export options.