docker exec -it todoe-mongo mongosh -u root -p root --authenticationDatabase admin todoe
db.task_events.find().pretty()

docker exec -i todoe-mysql mysql -u todoe -ptodoe todoe_onboarding -e "SHOW TABLES;"
docker exec -i todoe-mysql mysql -u todoe -ptodoe todoe_onboarding -e "DROP TABLE IF EXISTS users_events;"
docker exec -i todoe-mysql mysql -u todoe -ptodoe todoe_onboarding -e "DROP TABLE IF EXISTS users_view;"
docker exec -i todoe-mysql mysql -u todoe -ptodoe todoe_onboarding -e "SHOW TABLES; SELECT * FROM schema_migrations;"
docker exec -i todoe-mysql mysql -u todoe -ptodoe todoe_onboarding -e "SELECT * FROM users_view;"

1. Roll back N migrations cleanly (preferred):

Use the migrate CLI — it updates schema_migrations for you. No manual SQL.

go run -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1 \
    -path domain/user/adapter/migrations \
    -database "mysql://todoe:todoe@tcp(localhost:3306)/todoe_onboarding?multiStatements=true" \
    down 1                # or "down" alone for everything

3. Recover from "dirty" state:

If a migration crashed mid-file, dirty=1 is set and the runner refuses to retry. Two ways out:

# (a) Manually fix the DB to whatever state version N intends, then mark clean:
go run -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1 \
-path domain/user/adapter/migrations \
-database "mysql://...?multiStatements=true" \
force 0               # treat DB as if no migrations applied
# then re-run onboarding to apply forward.

# (b) Or just nuclear-reset (option 2) — same outcome with less thinking.

Rule of thumb: never UPDATE schema_migrations SET dirty=0 directly. Either use migrate down / migrate force, or wipe the table along with the data
tables. Hand-editing the bookkeeping while leaving the data half-applied is how you end up with skew between what the runner thinks is applied and
what's actually in the DB.