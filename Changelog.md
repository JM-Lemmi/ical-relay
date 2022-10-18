# v1.1.6

- add move-time to `edit-bysummary-regex`-module

# v1.1.5

- ignore unavailible URLs

# v1.1.4

- add remote calendar URL to Debug log

# v1.1.3

- remove view handle and replace with frereit/reacht-calendar

# v1.1.2

- Module edit-byid and edit-byregex:
  - Hotfix for [#39](https://www.github.com/JM-Lemmi/ical-relay/issues/39): Empty property will no be filled in "overwrite" mode.

# v1.1.1

- `edit-byid` & `edit-bysummary-regex` now have an `overwrite` parameter.
- improve logs:
  - Now shows profile in every logmessage of handler
  - Recognises X-Forwarded-For as Client IP
  - log output validation in info

# v1.1.0

- Add `save-to-file`-Module
- `delete-timeframe`-Module now accepts "now" as valid value for before and after
- empty source is now allowed and create a new empty calendar