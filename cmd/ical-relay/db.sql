CREATE TABLE IF NOT EXISTS schema_upgrades (
    version integer NOT NULL,
    date_installed TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS source (
    url text NOT NULL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS profile (
    name           text NOT NULL PRIMARY KEY,
    public         bool NOT NULL,
    immutable_past bool NOT NULL
);

CREATE TABLE IF NOT EXISTS profile_sources (
    profile text REFERENCES profile(name) ON DELETE CASCADE NOT NULL,
    source  text REFERENCES source(url) ON DELETE CASCADE NOT NULL,
    UNIQUE  (profile, source)
);

CREATE TABLE IF NOT EXISTS action (
    id         integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    type       text NOT NULL,
    parameters jsonb NOT NULL
);

/* TODO: add facility for easy reordering */
CREATE TABLE IF NOT EXISTS rule (
    id         integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    profile    text references profile(name) NOT NULL,
    operator   text NOT NULL,
    action     integer REFERENCES action(id) ON DELETE CASCADE NOT NULL,
    expiry     timestamp with time zone
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
    note    text
);

CREATE TABLE IF NOT EXISTS notifier (
    name     text NOT NULL PRIMARY KEY,
    UNIQUE   (source, interval),
    source   text NOT NULL,
    interval interval HOUR TO SECOND NOT NULL
);

CREATE TABLE IF NOT EXISTS recipients (
    email text PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS notifier_recipients (
    notifier  text REFERENCES notifier(name) ON DELETE CASCADE NOT NULL,
    recipient text REFERENCES recipients(email) ON DELETE CASCADE NOT NULL,
    UNIQUE    (notifier, recipient)
);