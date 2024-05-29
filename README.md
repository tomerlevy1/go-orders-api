# An Order API microservice written in Golang

## Notes

### Modules that we'll use

- [go-chi/chi/v5](github.com/go-chi/chi/v5)
 
### Install modules

One approach to do so is to use the `go get` command. For example, to install the `gorilla/mux` module, run the following command:
```bash
go get -u github.com/gorilla/mux
```

Another approach is to import the module somewhere in the code. For example, to import the `gorilla/mux` module, add the following line to the code:
```go
import "github.com/gorilla/mux"
```
And then run the following command:
```bash
go mod tidy
```

