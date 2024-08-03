DROP TABLE IF EXISTS module;

ALTER TABLE profile DROP COLUMN IF EXISTS source;

CREATE TABLE IF NOT EXISTS source (
    url text NOT NULL PRIMARY KEY
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