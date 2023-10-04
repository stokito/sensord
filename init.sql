CREATE SCHEMA sensors;

SET
    search_path TO sensors;


-- период, сумма_всех_показаний, количество_показаний, среднее, минимальное, максимальное
CREATE TABLE measurement
(
    period_start TIMESTAMP        NOT NULL,
    sensor_id    INT              NOT NULL,
    total_count  BIGINT           NOT NULL,
    total_sum    DOUBLE PRECISION NOT NULL,
    agv_value    DOUBLE PRECISION NOT NULL,
    min_value    DOUBLE PRECISION NOT NULL,
    max_value    DOUBLE PRECISION NOT NULL
);

CREATE UNIQUE INDEX idx_measurement
    ON measurement (period_start, sensor_id) INCLUDE (total_count, total_sum, agv_value, min_value, max_value);
