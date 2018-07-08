CREATE TABLE core (
    name text,
    description text,
    config text,
    api_key text,
    api_secret text,
    style text,
    footer text,
    domain text,
    version text,
    migration_id integer default 0,
    use_cdn bool default false
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR (50) UNIQUE,
    password text,
    email VARCHAR (50) UNIQUE,
    api_key text,
    api_secret text,
    administrator bool,
    created_at TIMESTAMP,
    pushover_key text default ''
);

CREATE TABLE services (
    id SERIAL PRIMARY KEY,
    name text,
    domain text,
    check_type text,
    method text,
    port integer,
    expected text,
    expected_status integer,
    check_interval integer,
    post_data text,
    order_id integer,
    created_at TIMESTAMP
);

CREATE TABLE hits (
    id SERIAL PRIMARY KEY,
    service INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE ON UPDATE CASCADE,
    latency float,
    created_at TIMESTAMP WITHOUT TIME zone
);

CREATE TABLE failures (
    id SERIAL PRIMARY KEY,
    issue text,
    method text,
    service INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMP WITHOUT TIME zone
);

CREATE TABLE checkins (
    id SERIAL PRIMARY KEY,
    service INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE ON UPDATE CASCADE,
    check_interval integer,
    api text,
    created_at TIMESTAMP
);

CREATE TABLE communication (
    id SERIAL PRIMARY KEY,
    method text,
    host text,
    port integer,
    username text,
    password text,
    var1 text,
    var2 text,
    api_key text,
    api_secret text,
    enabled boolean,
    removable boolean,
    limits integer,
    created_at TIMESTAMP
);

CREATE INDEX idx_hits ON hits(service);
CREATE INDEX idx_failures ON failures(service);
CREATE INDEX idx_checkins ON checkins(service);