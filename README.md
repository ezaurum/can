# Redis repository

# Test
```shell
docker-compose up -d
go test .
docker run --network="host" -it --rm redis redis-cli -h localhost -p 13789
```
