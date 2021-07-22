# tiny-redis-cache
Проект представляет собой реализацию прототипа in-memory хранилища, доступ к которому осуществляется через REST API.  
Поддерживаемые операции: SET(с возможностью установки Time To Live), GET, DEL.  
На данный момент пока не успел реализовать операцию KEYS.  

Сборка и запуск кэш-сервера:

```
go build -o tmp/cache-server cmd/apiserver.go
./tmp/cache-server <port>
```

Клиентская библиотека находится в /api/client, запуск примера использования (необходимо сначала запустить сервер):

```
go build -o tmp/client_example cmd/apiclient_example.go
./tmp/client_example localhost:<server-port>
```

---

Написаны тесты для компонент storage: `go test -v -race ./storage`  
и server: `go test -v ./api/server`.  
