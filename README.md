## Rocketbase

> Rocketbase = Pocketbase + PostgreSQL + a lot of goodies

## Setup

```bash
psql
```

```sql
CREATE
USER your_user WITH PASSWORD 'your_pass';
CREATE
DATABASE rocketbase;
CREATE
DATABASE rocketbase_logs;
GRANT ALL PRIVILEGES ON DATABASE
rocketbase TO your_user;
GRANT ALL PRIVILEGES ON DATABASE
rocketbase_logs TO your_user;
-- test
CREATE
DATABASE rocketbase_test;
CREATE
DATABASE rocketbase_logs_test;
GRANT ALL PRIVILEGES ON DATABASE
rocketbase_test TO your_user;
GRANT ALL PRIVILEGES ON DATABASE
rocketbase_logs_test TO your_user;
```

## Credit

- pocketbase for main codebase
- postgresbase for adapting postgres

