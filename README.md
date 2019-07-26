# GRPC microservices

The repository includes two microservices. A gRPC server, which is made in Python and gRPC client, which is made in Go.

It is a demonstration of how to apply a discount to the products with microservices.

## Install gRPC server

```
pip install poetry
cd discount
poetry install
poetry shell
python server.py 11443
```

## Install gRPC client

```
cd catalog
go run main.go 4000
```

Once you run the client, you can fetch data with the following command:

```
curl http://localhost:4000/products/1
```
