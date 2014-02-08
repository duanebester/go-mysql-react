# Post User
curl -v -H "Accept: application/json" -H "Content-type: application/json" -H "Authorization: xY57efl9c8lNUMkXES8RGQ" -X POST -d '{"name":"Duane","last":"Bester","email":"duane@snowdrop.io","password":"password"}'  http://localhost:8080/api/user


# Post Agent
curl -v -H "Accept: application/json" -H "Content-type: application/json" -H "Authorization: xY57efl9c8lNUMkXES8RGQ" -X POST -d '{"name":"MontyA","userid":1}'  http://localhost:8080/api/agent


# Post Alert
curl -v -H "Accept: application/json" -H "Content-type: application/json" -H "Authorization: xY57efl9c8lNUMkXES8RGQ" -X POST -d '{"message":"Hello","level":1,"category":"database"}'  http://localhost:8080/api/alert