Go backend for the ToDo
==================

### How to start

```bash
go build
./todo-go
```


### DataBase 

Can work with MySQL or SQLite ( default )

- following is the config for MySQL:
```yaml
db:
  type: mysql
  user: root
  password: 1
  host: localhost
  database: todos
```
(you need to create the database (code will init all necessary tables on its own))

- following is the config for SQLite:
```yaml
db:
  type: sqlite
  path: db.sqlite
```
