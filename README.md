# nestql

nestql is a no-code tool that converts sql statements to restfull APIs.

Especially suitable for mock server development. 

You just need to provide the route paths and the associated SQL statements, your Web API server is ready!

A sample config file looks like this:

```yaml
{
  "dbConnString": "postgres://postgres:pass1234@localhost:5432/mydb",
  "services": [
    {
      "route": "/persons",
      "query": "select id, name, birthday from person"
    },
    {
      "route": "/person/:id",
      "query": "select id, name, dob from test2 where id = {{param.id}}"
    },
    {
      "route": "/departments",
      "query": "select id, name from department"
    }
  ]
}
```


curl http://localhost:8080/persons

response:
```yaml
[
    {
        "id": 1,
        "name": "hakan"
        "birthday": "1995-08-12T00:00:00Z",
    },
    {
        "id": 2,
        "name": "george"
        "birthday": "2011-07-11T00:00:00Z",
    },
    {
        "id": 3,
        "name": "helen"
        "birthday": "1982-05-04T00:00:00Z",
    }
]
```


curl http://localhost:8080/person/2

response:
```yaml
  {
        "id": 2,
        "name": "george"
        "birthday": "2011-07-11T00:00:00Z",
  }
 
```

