
## Run postgresql docker container.
`bash
docker run --rm --name local-postgres -v /home/maxx/dev/golang/greenlight/.postgres-data:/var/lib/postgresql/data -p 5431:5432 -d postgres
`
