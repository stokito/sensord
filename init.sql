CREATE DATABASE sensors;

CREATE SCHEMA sensors;

SET search_path TO sensors;

CREATE TABLE measurements
(
    measure_time TIMESTAMP NOT NULL,
    sensor_id    INT       NOT NULL REFERENCES sensors(id),
    value        DOUBLE PRECISION
);

CREATE INDEX idx_measurements
    ON measurements (measure_time, sensor_id)
    INCLUDE (value);



CREATE TABLE sensors
(
    id    INT
        CONSTRAINT sensors_pk PRIMARY KEY,
    name  VARCHAR DEFAULT '' NOT NULL,
    room  VARCHAR DEFAULT '' NOT NULL,
    props JSONB
);

