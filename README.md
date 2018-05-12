# A TCP Server in Go
Sample code for building a toy TCP server with sending a query to an external weather API ([APIXU](https://www.apixu.com)).

## Features:
* Accept multiple connections.
* Throttle request rate to external weather API
* Statistics in HTTP server
* Network error handling as well

## Setup:
* Copy the configuration and replace the `WEATHER_API_KEY` with your own:

    ```
    cp config.go.example config.go
    ```

* Run the server:

    ```
    go run *.go
    ```
* To test your server, send some raw data to that port:

    ```
    nc localhost 3333
    ```

## Usages:
TCP server takes in any request text per line. The commands are listed below:

* **weather,[city]** *(Temperature of the city, ex: weather,taipei)*
* **time** *(Local time in the server)*
* **status** *(statistics report in plain text)*
* **quit** *(Quit the TCP connection)*
* **help** *(Hints)*


