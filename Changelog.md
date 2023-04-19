# v2.0.0-beta.5

- Replace Modules with Filters and Actions
  - Filters: Regex, Id, Timeframe, Duplicates, All, Duration
  - Actions: Delete, Edit, AddReminder, StripInfo
- Multiple Sources per Profile
  - base64, profile, file and http/https
- Configuration upgrade for legacy configurations
- Module instances are now called "Rules"
- Frontend:
  - "Today" marker
  - clickable location links
  - Disabled Edit Button for past
  - Copy ICS Link to Clipboard
  - Configurable Name and Favicon
- Api to add new Calendar Entries

# v2.0.0-beta.4

- Templates are now in `/opt/ical-relay/templates` by default and can be changed by config setting.
- Use CDN for javascript
- Frontend
  - Add Year
  - Show Error if Token is invalid
  - Hide Edit Button for Past, when immutable past is active
  - Automatic Redirection to saved Profile
  - Sort Events
  - User How-To
  - Configurable Dataprivacypolicy and Impressum Links

# v2.0.0-beta.3.2

- fix navbar subscribe link
- fix html lang
- add Delete Button functionality

# v2.0.0-beta.3.1

- Fix: relative path for static assets #89
- add Selector to Index Page
- add ICS link to navbar

# v2.0.0-beta.3

- Frontend
  - Module Hinzufügen oder Entfernen
  - Mail-Benachrichtigungen hinzufügen oder entfernen
- Fix templates folder in docker & debian package

# v2.0.0-beta.1

- API
  - add authentication (admin and superadmin roles)
  - Single Calendar Entry: `/api/profiles/{profile}/calentry` POST and DELETE
  - Modules: `/api/profiles/{profile}/modules` POST
- Add Frontend
  - Monthly view
  - Edit view for Single Entries

# v1.3.1

- basic RRULE handling in delete-timeframe Module
  - waiting for upstream RRULE Handling in golang-ical
  - cannot handle COUNT
  - cannot handle when timeframe is inbetween or only in the beginning of the RRULE
- fix "invalid start time" in delete-timeframe with only "before"-option

# v1.3.0

- Immutable past is now a profile boolean option. Simply add `immutable-past: true` to your profile configuration.
- Query Parameters can be used to start a module with parameters given at runtime.
  - `?reminder=15m` adds an Alarm to every Entry of 15 minutes before. It is not supported in dynamic calendars from Outlook or Google.
- Notification Mails now look much more readable and don't deliver the whole ICS Event.

# v1.2.0

- add Notifiers: Get Notifications per Mail, if a calendar changes
  - periodic
  - from cronjobs with `--notifier`
- add API
  - `/api/calendars`: Returns all Public Calendars as json-array.
  - `/api/reloadconfig`: Reloads the config from disk.
  - `/api/notifier/<notifier>/addrecipient`: with an E-Mail Address as body adds the recipient to the notifier.
- Release as `.deb` Package

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