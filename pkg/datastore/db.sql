CREATE TABLE IF NOT EXISTS schema_upgrades (
    version integer NOT NULL,
    date_installed TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS source (
    id  integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    url text NOT NULL
);

CREATE TABLE IF NOT EXISTS profile (
    name           text NOT NULL PRIMARY KEY,
    public         bool NOT NULL,
    immutable_past bool NOT NULL
);

CREATE TABLE IF NOT EXISTS profile_sources (
    profile text    REFERENCES profile(name) ON DELETE CASCADE NOT NULL,
    source  integer REFERENCES source(id) ON DELETE CASCADE NOT NULL,
    UNIQUE  (profile, source)
);

/* TODO: add facility for easy reordering */
CREATE TABLE IF NOT EXISTS rule (
    id                integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    profile           text references profile(name) NOT NULL,
    operator          text NOT NULL,
    action_type       text NOT NULL,
    action_parameters jsonb NOT NULL,
    expiry            timestamp with time zone
);

CREATE TABLE IF NOT EXISTS filter (
    id         integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    rule       integer REFERENCES rule(id) ON DELETE CASCADE NOT NULL,
    type       text NOT NULL,
    parameters jsonb NOT NULL
);

CREATE TABLE IF NOT EXISTS admin_tokens (
    profile text REFERENCES profile(name) ON DELETE CASCADE NOT NULL,
    token   char(64) PRIMARY KEY,
    note    text /* TODO: Possibly add NOT NULL and just save an empty string? */
);

CREATE TABLE IF NOT EXISTS notifier (
    name     text NOT NULL PRIMARY KEY,
    source   text NOT NULL
);

CREATE TABLE IF NOT EXISTS recipients (
    recipient     text PRIMARY KEY,
    type            text NOT NULL
);

CREATE TABLE IF NOT EXISTS notifier_recipients (
    notifier  text REFERENCES notifier(name) ON DELETE CASCADE NOT NULL,
    recipient text REFERENCES recipients(recipient) ON DELETE CASCADE NOT NULL,
    UNIQUE    (notifier, recipient)
);

CREATE TABLE IF NOT EXISTS notifier_history (
    id          integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    notifier    text REFERENCES notifier(name) ON DELETE CASCADE NOT NULL,
    recipient   text REFERENCES recipients(recipient) ON DELETE CASCADE NOT NULL,
    type        TEXT NOT NULL,
    eventDate   TIMESTAMP WITH TIME ZONE NOT NULL,
    modifyDate  TIMESTAMP WITH TIME ZONE NOT NULL,
    data        JSON NOT NULL
)