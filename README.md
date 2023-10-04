# sensord Sensors Daemon: collect metrics from sensors

## Engineer assignment
### Intro
1. Meet Yochbad, she would like to collect temperature samples from multiple sensors
   spread across the `Hagag` building in order to check the temperature fluctuations in the
   building.
2. She has multiple sensors that send those measurements to a server, during the day
   randomly.
3. She wants to have the max, min and average temperature stored for each day, for each
   sensor for the past week.
4. In addition, she wants to have a min, max and avg metric for the past week for every
   sensor and for all sensors combined.

###  Description
1. Create a workable design on how to implement a server that can handle such sensor
   data, store it and present the information to the user.
2. Implement a server that will be able to handle hundreds sensors.
3. Implement an ability to present these settings (simple!) using simple commands to the
   screen or to a file:
   - Show the max, min, avg temps for every sensor.
   - Show the max, min, avg temps for all sensors.
4. Make sure you use the proper data structure that can handle this data efficiently.
5. Don’t forget to consider cases where several sensors may try to send the results in
   parallel, moreover, some might reside in super fluctuating environments and send every
   second.


## Implementation

Here for the task we can use a specialized time series DB like InfluxDB or Prometheus.
But for a simplicity I will use a PostgreSQL which is a swiss army knife of DBs.
For the best performance I will use FastHttp library which is designed to have minimal memory allocations.

Database has a table `measurement` that collects measurements stats in aggregated form:

* `measurement_day` Day for which we collect stats
* `sensor_id`
* `total_count` total received measurements
* `total_sum` Sum of all values e.g. temperature
* `avg_value` Average temperature
* `min_value` Minimal temperature
* `max_value` Maximal temperature

The stats records are updated with upsert for a pair of day and sensor.
The default PostgreSQL read-committed isolation level allows to safely parallel update records.

## Configuration

You need to configure environment variables:
* `SENSOR_LISTEN_HTTP` Sensor HTTP API listen address. You can specify `hostname:port` or just `:port`
* `ADMIN_LISTEN_HTTP` Admin HTTP API listen address
* `DB_URL` PostgreSQL database URL. E.g. `postgres://postgres:postgres@localhost:5432/sensorsdb?sslmode=disable&search_path=sensors`
* `DB_LOG` if `true` then log SQL queries and args. Useful for testing and debug.

See the .env file with example for a local running.

## Running

   docker-compose up

This will start locally a PostgreSQL, create a DB and start the sensord API.

You can build manually the sensord with:

   go build -o /build/sensord ./cmd/sensord/main.go


The sensord has a Dockerfile and you can build an image with:

    docker build -t sensord .

The Dockerfile uses two stage build.

## Testing
Most logic is on the DB layer so see the db_pg_test.go
The test will start a PostgreSQL server in a docker container.

## API endpoints
The API is separated into two parts:
* Sensor API for sensors
    * `POST http://localhost:8080/api/v1/measurement` receives a JSON with measurements.
* Admin API for Yochbad so she can watch reports
    * `GET http://localhost:9090/api/v1/stats/Total` aggregated data for last week for all sensors.
    * `GET http://localhost:9090/api/v1/stats/EachSensor` report by each sensor for last week e.g. today's midnight minus 7 days.
    * `GET http://localhost:9090/api/v1/stats/EachSensorAndDay` report grouped by each sensor and a day.

Having this two API separated allows to secure them with different way.
For example the Sensor API may use plain HTTP and have no authorization.
Or use mutual TLS between the sensor and sensord.
But the Admin API may use TLS, Basic Auth and be accessible only from specific IP.

Since the Sensor API has a big load it's based on FastHttp.

```sh
curl -X POST --location "http://localhost:8080/api/v1/measurement" \
-H "Content-Type: application/json" \
-d "{
\"sensorId\": 1,
\"time\": \"2023-01-02T00:00:00.000Z\",
\"value\": 42
}"
```