# Redis repository

# Test
```shell
docker-compose up -d
go test .
docker run --network="dev_can_default" -it --rm redis redis-cli -h redis 
```
