# Post User
curl -v -H "Accept: application/json" -H "Content-type: application/json" -H "Authorization: -gnVfF6Es_YmPWpC6mMoGw" -X POST -d '{"name":"Duane","last":"Bester","email":"duane@snowdrop.io","password":"password"}'  http://localhost:8080/api/user


# Post Agent
curl -v -H "Accept: application/json" -H "Content-type: application/json" -H "Authorization: -gnVfF6Es_YmPWpC6mMoGw" -X POST -d '{"name":"MontyA","userid":1}'  http://localhost:8080/api/agent


# Post Alert
curl -v -H "Accept: application/json" -H "Content-type: application/json" -H "Authorization: -gnVfF6Es_YmPWpC6mMoGw" -X POST -d '{"message":"Hello","level":1,"category":"database"}'  http://localhost:8080/api/alert


# Get Alerts
curl -v -H "Accept: application/json" -H "Content-type: application/json" -H "Authorization: h_IOwaHJZvSL8kYNkQViJg" -X GET http://localhost:8080/api/alerts
