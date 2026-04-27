podman exec -it todoe-mongo mongosh -u root -p root --authenticationDatabase admin todoe


db.task_events.find().pretty()