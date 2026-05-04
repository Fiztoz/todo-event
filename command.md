docker exec -it todoe-mongo mongosh -u root -p root --authenticationDatabase admin todoe
db.task_events.find().pretty()

docker exec -i todoe-mysql mysql -u todoe -ptodoe todoe_onboarding -e "SHOW TABLES;"
docker exec -i todoe-mysql mysql -u todoe -ptodoe todoe_onboarding -e "DROP TABLE IF EXISTS users_events;"
docker exec -i todoe-mysql mysql -u todoe -ptodoe todoe_onboarding -e "DROP TABLE IF EXISTS users_view;"
docker exec -i todoe-mysql mysql -u todoe -ptodoe todoe_onboarding -e "SHOW TABLES; SELECT * FROM schema_migrations;"

  1. Roll back N migrations cleanly (preferred):

  Use the migrate CLI — it updates schema_migrations for you. No manual SQL.

  go run -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1 \
    -path domain/user/adapter/migrations \
    -database "mysql://todoe:todoe@tcp(localhost:3306)/todoe_onboarding?multiStatements=true" \
    down 1                # or "down" alone for everything
