curl -X POST http://localhost:3000/tasks -H "Content-Type: application/json" -d '{"title":"my new task"}'

curl -X PATCH http://localhost:3000/tasks/69f318830fb7bff0f0f7b3fd/status -H "Content-Type: application/json" -d '{"status":"in_progress"}'