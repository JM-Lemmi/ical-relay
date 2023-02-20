CREATE TABLE IF NOT EXISTS schema_upgrades (
    version integer NOT NULL,
    date_installed TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS profile (
    name           text NOT NULL PRIMARY KEY,
    source         text NOT NULL,
    public         bool NOT NULL,
    immutable_past bool NOT NULL
);

CREATE TABLE IF NOT EXISTS module (
    id         integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    profile    text references profile(name) NOT NULL,
    name       text NOT NULL,
    parameters jsonb NOT NULL
);

CREATE TABLE IF NOT EXISTS admin_tokens (
    profile text REFERENCES profile(name) ON DELETE CASCADE NOT NULL,
    token   char(64) PRIMARY KEY
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