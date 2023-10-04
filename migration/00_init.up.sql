CREATE SCHEMA sensors;

SET
    search_path TO sensors;


CREATE TABLE measurement
(
    measurement_day DATE             NOT NULL,
    sensor_id       INT              NOT NULL,
    total_count     BIGINT           NOT NULL,
    total_sum       DOUBLE PRECISION NOT NULL,
    avg_value       DOUBLE PRECISION NOT NULL,
    min_value       DOUBLE PRECISION NOT NULL,
    max_value       DOUBLE PRECISION NOT NULL
);

CREATE UNIQUE INDEX idx_measurement
    ON measurement (measurement_day, sensor_id)
    INCLUDE (total_count, total_sum, avg_value, min_value, max_value);
