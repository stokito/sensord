CREATE SCHEMA sensors;

SET
search_path TO sensors;

CREATE TABLE sensors
(
    id    INT
        CONSTRAINT sensors_pk PRIMARY KEY,
    name  VARCHAR DEFAULT '' NOT NULL,
    room  VARCHAR DEFAULT '' NOT NULL,
    props JSONB
);

-- период, сумма_всех_показаний, количество_показаний, среднее, минимальное, максимальное
CREATE TABLE measurement
(
    measure_start TIMESTAMP        NOT NULL,
    sensor_id     INT              NOT NULL REFERENCES sensors (id),
    total_count   BIGINT           NOT NULL,
    total_sum     DOUBLE PRECISION NOT NULL,
    agv_sum       DOUBLE PRECISION NOT NULL,
    min_value     DOUBLE PRECISION NOT NULL,
    max_value     DOUBLE PRECISION NOT NULL
);

CREATE UNIQUE INDEX idx_measurement
    ON measurement (measure_start, sensor_id) INCLUDE (total_count, total_sum, agv_sum, min_value, max_value);
