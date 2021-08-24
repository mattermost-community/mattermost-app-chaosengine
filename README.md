# Chaos Engine App
ChaosEngine app nominate randomly a different MoD and On-call person for every Gameday. The On-Call person is different from the normal process, so the team can practice and act independently when the time comes.

The Chaos Engine App is integrated in Mattermost and you can create:
- Chaos Teams create `/chaos-engine team create --name sre --member @spiros`
- Chaos Teams list `/chaos-engine team list`
- Chaos Gamedays create `/chaos-engine gameday create --name chaos-august --team sre --schedule-at 2021-25-08 07:00:00` (pending feature)
- Chaos Gameday Start `/chaos-engine gameday start --day chaos-august` (pending feature)
- Chaos Gameday list `/chaos-engine gameday list` (pending feature)

## Running

Here are available configuration to run the app:

| Name                  | Default                           | Description |
|-----------------------|-----------------------------------|---------|
| debug                 | false                             | debug logging |
| address               | :3000                             | listening address |
| app.type              | http                              | mattermost app type |
| app.root_url          | http://localhost:3000             | the root url of the app |
| db.scheme             | http://localhost:3000             | the scheme, supports `sqlite3`, `postgres`, `postgresql`|
| db.url                | sqlite3://engine.db"              | the database URL which can be sqlite DB or Postgres DSN |
| db.idle_conns         | 2                                 | the number of idle connections |
| db.max_open_conns     | 1                                 | the max number of open connections |
| db.max_conn_lifetime  | 1                                 | the max connection lifetime |


Run the server:
```
make run
```

By default the application if you run will use `sqlite3` database (for local dev only) and will listen to `:3000` and the root url
will be `http://localhost:3000`

### Overriding defaults

There are two ways to override the defaults.
- `config.yml` file only for local development
- environment variables with prefix `CHAOS_ENGINE_*`

An example with `config.yml` for local dev only:

```yaml
debug: true
environment: prod
address: ":3001"
# mattermost app config
app:
 type: http
 root_url: "http://localhost:3000"
# database config
db:
 scheme: sqlite3
 url: "sqlite3://mytestdb.db"
```

An example with environment variables:
```bash
CHAOS_ENGINE_DEBUG=true
CHAOS_ENGINE_ADDRES=":3000"
CHAOS_ENGINE_APP_TYPE="http"
CHAOS_ENGINE_DB_SCHEME="sqlite3"
```

### Lint

Linting the codebase:
```
make lint
```

### Test

Run the tests:
```
make test
```

## DEV env with Mattermost

You can run this [`docker-compose`](https://github.com/mattermost/mattermost-plugin-apps/tree/master/dev) to run
Mattermost and install your app with the following `slash` command.

`/apps install http http://localhost:3000/manifest`
