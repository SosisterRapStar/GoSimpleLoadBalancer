# How to run 
Open /loadbalancer

## Docker
* Use your config and mount it in docker-compose
* Run docker-compose 

## Local
* Use go install or go build to compose binary
* Create a *config.yaml* or *<your_file_name>.yaml* and save its path to `BALANCER_CONFIG_PATH` environment variable 
* Start binary

## Config file
This is the common look of the config file

```yaml
listen: "0.0.0.0:3002"

balance: "random"

healthcheck:
  period: 10
  responseTimeoutSeconds: 10
  maxRetries: 3
  timeOutStep: 2
  endpoint: "/health"

upstreams:
  - "localhost:8000"
  - "localhost:8001"
```
### Listen field
Here you can specify the host and port for loadbalancer

### Balance field
You can use *roundrobin* or *random* balance algos. If field is not specified default algo is *roundrobin*

### Upstreams field
Use upstreams to configure backends addreses

### Healthcheck field
Use params for healthcheck here. You can also omit all the fields except *endpoint* this field is necessary.


## Logs
Balancer pushes logs to stdout so tou can use
* **docker logs <cont_name>** to see logs