CREATE TABLE IF NOT EXISTS metrics
(
    id serial PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    CONSTRAINT uniq_name_type UNIQUE (name,type)
);