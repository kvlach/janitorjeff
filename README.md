# JanitorJeff

[![Go Documentation](https://godocs.io/git.sr.ht/~slowtyper/janitorjeff?status.svg)](https://godocs.io/git.sr.ht/~slowtyper/janitorjeff)

A general purpose, cross-platform bot.

*Very much under active development, there will most likely be breaking changes.*

## Running

You can either use the provided `docker-compose.yml` or if you don't want to use
docker:

### PostgreSQL

Install PostgreSQL 15.
To initialize the database, enter the PostgreSQL interactive terminal
by executing `psql` and then run:

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

Make sure redis is installed and afterward run:

```sh
redis-server redis.conf --daemonize yes
```

### Prometheus
```sh
prometheus --config.file=prometheus.yml
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
export PROMETHEUS_ADDR=localhost:2112
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

## Contributing

### Directory Structure
- `core/`: The juicy part, all the core interfaces, structs and functions are located here.
- `frontends/`: The annoying frontend specific glue-code.
- `commands/`: Implementations of all the commands, the bulk of the code exists here.

### Scope
One of the core ideas behind Jeff is that he supports multiple frontends.
To achieve this in a sensible way, i.e. not having to implement each command for
each frontend, there must be abstractions, the main of which is the "scope".
A scope can be anything, a Discord user, a Discord channel, a Twitch channel, etc.;
these are all scopes which are assigned a unique number by *Jeff*.
This unique ID is only used internally.

#### Person, Author, Place, Here
Within the code, these 4 names are often used.
They are all scopes, there is no difference in, for example,
how they are stored in the database (all exist in the `scopes` table),
but this naming convention is used to make it clear what the scope in each instance represents:

| Term   | Definition                                                    |
|--------|---------------------------------------------------------------|
| Person | A `scope` which refers to a `person` (users).                 |
| Author | The `person` who authored the event.                          |
| Place  | A `scope` that refers to a `place` (servers, channels, etc.). |
| Here   | The `place` in which the event took place.                    |

### Command Types
- Normal
- Advanced
- Admin

You may find more information in `core/commands.go`.

### Events
You may find more information in `core/events.go`.

### Settings
You may find more information in `core/db.go`

### Errors
There are 2 types of errors: one concerns the developer (something
unexpected happened during execution) the other concerns the user (the user
did something incorrectly) and should thus be informed appropriately.
Naming convention:
- `err` = normal errors
- `urr` = user errors
