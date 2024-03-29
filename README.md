# User rating web crawler [![CI](https://github.com/Kowol/user-rating-crawler/actions/workflows/main.yaml/badge.svg?branch=main)](https://github.com/Kowol/user-rating-crawler/actions/workflows/main.yaml)

Web crawler for scrapping rating from Roku site. It's using headless chromium under the hood to render the full react page - that's the reason why it's a bit slow - but don't worry, we can speed it up by scaling up workers. 

## Solution description

App providers 2 separate services (two separate binaries - as they share part of codebase and share the domain). But they could be scaled individually

First one is a worker - it's simple RabbitMQ consumer that process the urls.

Second one is a GRPC client - it accepts urls to process and push them to the exchange

## Requirements

* Docker

## Setup

### Docker

To build all services required for this application please run following command

```shell
docker-compose up -d --build 
```

It will bring up all the required containers.

### MongoDB

MongoDB is accessibly under `localhost:27017` (login: `channelCrawler`, pass: `pass`). Auth db: `admin`. All indexed
channels could be found in `channel` collection - it would be created on first run of app

### Containers specification

* rabbitmq - AMQP queue - holds all the messages to process
* db - MongoDB database - holds all the results of our processing
* crawler-api - GRPC API responsible for collecting data to process. It exposes 2 endpoints (for single and batch
  requests)
* crawler-worker - AMQP Consumer that crawl URLs provided by the queue and saves them to database
* crawler-client - Simple client that allow to push CSV to the GRPC API

## Running the crawler

Prepare CSV file in following format. It can contain as many lines as you wish

| URL                                                                                 |
|-------------------------------------------------------------------------------------|
| https://channelstore.roku.com/details/96da35e0bce6c184b61e445cc6e62203/netflix      |
| https://channelstore.roku.com/details/afbca04cc0e1c93a2ea8f3382b56172c/prime-video  |
| https://channelstore.roku.com/details/d6ff1be180299e8be35ff79f5cc0628d/flickr  |

Run following command (but make sure that docker containers are up and running) or optionally
replace `$(pwd)/_examples/list.csv`
with path to your CSV file

```shell
docker-compose run -v $(pwd)/_examples/list.csv:/app/data.csv --rm crawler-client ./web-crawler-client --csv=data.csv
```

### Logs

Worker logs:

```shell
docker-compose logs crawler-worker
```

GRPC API logs:

```shell
docker-compose logs crawler-api
```

## Scaling

By default, consumer spawns 5 workers to work on messages - it could be changed via `CRAWLER_WORKERS_AMOUNT` env
variable.

Also, container could be scaled up to open new AMQP connections using following method:

```shell
docker-compose up -d --build --scale crawler-worker=<number of containers>
```

## Tests

### Unit tests

This command run only unit tests

```shell
make test-image
```

Or optionally `make test` - but it requires GO 1.18

### Integration and acceptance tests

This command run additional, heavy, integration and acceptance tests.

Integration test is covering only website scrapper using fake website (to be sure that elements are scrapped properly)

Acceptance test is covering end-to-end test on fake website (including communication with grpc, amqp and mongo)

All this stuff it done by using docker that runs the containers required for the tests

```shell
make test-integration-acceptance
```

## Additional information

### Env vars used

Whole configuration is done via env variables

| Name         | Description                                      | Default value   |
|--------------|--------------------------------------------------|-----------------|
| AMQP_URL | Full connection string to AMQP                   |                 |
| AMQP_QUEUE_NAME | Queue name                                       | channel_crawler |
| AMQP_EXCHANGE_NAME | Name od direct exchange                          | urls            |
| AMQP_ROUTING_KEY | Routing key used to route messages               | channel_url     |
| DATABASE_DSN | Database DSN                                     ||
| DATABASE_DB_NAME | Database name used for storing crawled data      | crawler         |
| GRPC_SERVER_PORT | GRPC API port                                    |                 |
| CRAWLER_WORKERS_AMOUNT | Amount of workers to spawn inside single process | 5               |


### TODO
The crawler itself could be polished with

* some random delay between calls
* faking user agent to prevent getting blocked
* maybe some caching
* other cool stuff

Also, the tracing part is missing - so it will require to generate some traceId with span to control the flow over the services.

RabbitMQ UI is available under http://localhost:15672/ (login: guest, password: guest) (here you can monitor all the
messages). Currently, DLX is not configured, so failed messages are removed. However, it's not a big deal to redeliver
failed messages from DLX

