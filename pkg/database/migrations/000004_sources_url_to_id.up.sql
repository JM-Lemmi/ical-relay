DROP TABLE profile_sources;
DROP TABLE source;

DROP TABLE filter, rule, action;

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


CREATE TABLE IF NOT EXISTS source (
    id  integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    url text NOT NULL
);

CREATE TABLE IF NOT EXISTS profile_sources (
    profile text    REFERENCES profile(name) ON DELETE CASCADE NOT NULL,
    source  integer REFERENCES source(id) ON DELETE CASCADE NOT NULL,
    UNIQUE  (profile, source)
);