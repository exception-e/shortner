# Укротитель ссылок 

### URL shortner

--------------
Pet project to study go
+ shortens entered url using murmur3 hash
+ rest endpoints:

```
GET http://localhost:8080/api/v1/shorten/{shortLink}

POST http://localhost:8080/api/v1/shorten
Content-Type: application/json
{
   "link":"https://google.com"
}
```
