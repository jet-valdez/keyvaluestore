## DOCKER
# Build
docker build --tag my-nginx .

# Run 
docker run --detach --publish 8080:80 --name nginx my-nginx



## CURL COMMANDS

curl -X PUT -d 'Hello, key-value store!' -v http://localhost:8080/v1/key/{key}

curl -v http://localhost:8080/v1/key/{key}