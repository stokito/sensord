# Engineer assignment
## Intro
1. Meet Yochbad, she would like to collect temperature samples from multiple sensors
   spread across the `Hagag` building in order to check the temperature fluctuations in the
   building.
2. She has multiple sensors that send those measurements to a server, during the day
   randomly.
3. She wants to have the max, min and average temperature stored for each day, for each
   sensor for the past week.
4. In addition, she wants to have a min, max and avg metric for the past week for every
   sensor and for all sensors combined.

##  Description
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

## Additional notes:
1. Make sure the code compiles and runs.
2. You have 3 hours, if you need more time - talk to me before.

## Implementation

Here for the task we can use a specialized time series DB like InfluxDB or Prometheus.
But for a simplicity I will use a PostgreSQL which is a swiss army knife of DBs.
For the best performance I will use FastHttp library which is designed to have minimal memory allocations.

We need to store measurements in aggregated form.


## Configuration

You need to configure environment variables:
* `LISTEN_HTTP` daemon listen address for HTTP API
