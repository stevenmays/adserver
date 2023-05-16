# Ad Server

This is a simple ad server written in Go that allows users to create ad campaigns, receive ad decisions, and track impressions.

## Getting Started

### Prerequisites

- Install [Go](https://golang.org/doc/install) (version 1.17 or higher)

### Running Tests

To run the tests, open a terminal and navigate to the project directory. Then, execute the following command:

```sh
go test -v
```

### Building the Project

To build the project, open a terminal and navigate to the project directory. Then, execute the following command:

```sh
go build -o adserver
```

### Running the Project

To run the project, execute the following command in the terminal:

```sh
./adserver
```

The server will start and listen on port 8000.

### Interacting with the API

You can use curl to interact with the API endpoints.

#### Create a new campaign

To create a new campaign, run the following curl command:

```sh
curl -i --header 'Content-Type: application/json' http://localhost:8000/campaign
--data-raw
'{"start_timestamp":1642493821,"end_timestamp":1742580221,"target_keywords":
["iphone","5G"],"max_impression":1,"cpm":20}'
```

#### Get an ad decision

To get an ad decision, run the following curl command:

```sh
curl -i --header 'Content-Type: application/json'
http://localhost:8000/addecision --data-raw '{"keywords":["iphone"]}'
```

#### Track an impression

To track an impression, replace <impression_id> with a valid impression ID and run the following curl command:

```sh
curl -X GET http://localhost:8000/<impression_id>
```
