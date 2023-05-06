# JanitorJeff

[![Go Documentation](https://godocs.io/git.sr.ht/~slowtyper/janitorjeff?status.svg)](https://godocs.io/git.sr.ht/~slowtyper/janitorjeff)

A general purpose, cross-platform bot.

*Very much under active developement, there will most likely be breaking changes.*

## Running

You can either use the provided `docker-compose.yml` or if you don't want to use
docker:

### PostgreSQL

Install PostgreSQL 15. To initialize the database enter the PostgreSQL
interactive terminal by executing `psql` and then run:

```sql
CREATE DATABASE jeff_db;
CREATE USER jeff_user WITH ENCRYPTED PASSWORD 'jeff_pass';
GRANT ALL PRIVILEGES ON DATABASE jeff_db TO jeff_user;
\c jeff_db postgres
-- You are now connected to database "jeff_db" as user "postgres".
GRANT ALL ON SCHEMA public TO jeff_user;
```

Exit the interactive terminal by executing `exit` and then run:

```sh
psql -U jeff_user -f schema.sql
```

### Redis

Make sure redis is installed and afterwards run:

```sh
redis-server redis.conf --daemonize yes
```

### Jeff

Make sure you are in the root of the project and run:

```sh
mkdir data/  # this directory is gitignored
```

Then `cd` into `data/` and create a `secrets.sh` file with the following
contents:

```sh
#!/bin/sh

export VIRTUAL_HOST=localhost
export PORT=5000
# DB
export POSTGRES_USER=jeff_user
export POSTGRES_PASSWORD=jeff_pass
export POSTGRES_DB=jeff_db
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_SSLMODE=disable
export REDIS_ADDR=localhost:6379
# Frontends
export DISCORD_TOKEN=token
export DISCORD_ADMINS=comma,seperated,list,of,user,ids
export TWITCH_ADMINS=comma,seperated,list,of,user,ids
export TWITCH_CHANNELS=comma,seperated,list,of,channel,names
export TWITCH_CLIENT_ID=cliend-id
export TWITCH_CLIENT_SECRET=client-secret
export TWITCH_OAUTH=oauth-token
# Commands
export MIN_GOD_INTERVAL_SECONDS=600
export OPENAI_KEY=api-key
export TIKTOK_SESSION_ID=session-id
export YOUTUBE=token
```

Finally, you can run Jeff by executing:

```sh
source data/secrets.sh && go run main.go -debug
```
