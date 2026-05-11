## Blog gRPC API

gRPC service with HTTP/JSON gateway, backed by PostgreSQL.

## Run

```bash
docker compose up --build
```

Server: `http://localhost:8080`  
gRPC: `localhost:50051`

## API

```bash
# create post
curl -X POST http://localhost:8080/v1/posts \
  -H "Content-Type: application/json" \
  -H "x-user-id: user1" \
  -d '{"body":"hello world"}'

# list posts
curl http://localhost:8080/v1/posts

# update post
curl -X PUT http://localhost:8080/v1/posts/<id> \
  -H "Content-Type: application/json" \
  -d '{"body":"updated"}'

# toggle like
curl -X POST http://localhost:8080/v1/posts/<id>/like \
  -H "x-user-id: user1"

# delete post
curl -X DELETE http://localhost:8080/v1/posts/<id>
```

User identity is passed via `x-user-id` header.

## Cache verification

```bash
# create a post
curl -s -X POST http://localhost:8080/v1/posts \
  -H "Content-Type: application/json" \
  -H "x-user-id: user1" \
  -d '{"body":"test"}' | jq .

# GET posts — cache miss, populates Redis
curl -s http://localhost:8080/v1/posts | jq .

# verify key exists in Redis
docker compose exec redis redis-cli KEYS "posts:*"

# GET posts again — cache hit (no DB query)
curl -s http://localhost:8080/v1/posts | jq .

# inspect cached value
docker compose exec redis redis-cli GET "posts:20:0"

# TTL remaining
docker compose exec redis redis-cli TTL "posts:20:0"
```

Create/update/delete post invalidates all `posts:*` keys. TTL: 60s.

## Development

### Regenerate proto

```bash
make gen
```

### Run without Docker

```bash
# start postgres
docker run -d --name blog-pg \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=blog \
  -p 5432:5432 postgres:16

go run ./cmd/server/main.go
```

`DATABASE_URL` env overrides the default DSN (`host=localhost user=postgres password=postgres dbname=blog port=5432 sslmode=disable`).