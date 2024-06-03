# An Order API microservice written in Golang

## TODO

- [ ] (TODO) Switch to GoDotEnv and use a `.env` file for the config
- [ ] (TODO) Move handler & model under the order package
- [ ] (TODO) Replace reference to the Repo with an interface
- [ ] (TODO) Support and switch to anoter data store (postgres)
- [ ] (TODO) Create E2E tests 

## Notes

### Modules that we'll use

- [go-chi/chi/v5](github.com/go-chi/chi/v5)
- [redis/go-redis](github.com/redis/go-redis)

### Install modules

One approach to do so is to use the `go get` command. For example, to install the `gorilla/mux` module, run the following command:
```bash
go get github.com/gorilla/mux
```

Another approach is to import the module somewhere in the code. For example, to import the `gorilla/mux` module, add the following line to the code:
```go
import "github.com/gorilla/mux"
```
And then run the following command:
```bash
go mod tidy
```

### Install redis

To install redis-cli, run the following command:
```bash
brew install redis
```

Then run it inside a docker.
```bash
docker run -p 6379:6379 redis:latest
```
