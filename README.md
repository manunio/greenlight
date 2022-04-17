# Greenlight golang Rest Api

## Run postgresql docker container
```bash
docker run --rm --name local-postgres -v /home/maxx/dev/golang/greenlight/.postgres-data:/var/lib/postgresql/data -p 5431:5432 -d postgres
```
`
## Create migration
```bash
migrate create -seq -ext=.sql -dir=./migrations create_user_table 
```

## Run migration
- ### up
  ```bash
    export GREENLIGHT_DB_DSN='postgres://greenlight:pa55word@localhost:5431/greenlight?sslmode=disable'
    migrate -path ./migrations -database $GREENLIGHT_DB_DSN up
  ```
- ### down
  ```bash
    export GREENLIGHT_DB_DSN='postgres://greenlight:pa55word@localhost:5431/greenlight?sslmode=disable'
    migrate -path ./migrations -database $GREENLIGHT_DB_DSN down
  ```
