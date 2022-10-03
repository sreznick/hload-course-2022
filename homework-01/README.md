
### Задание 1

Сделать микросервис коротких URL.

Поддерживаемые запросы

/create PUT

В теле запроса - JSON

* 'longurl': исходный URL


В ответе - JSON

* 'longurl': исходный URL
* 'tinyurl': короткий URL

Два запроса с одним исходным URL возвращают один и тот же короткий

Два запроса с разными исходным URL возвращают разные короткие

Размер tinyurl - 7 символов из алфавита a-ZA-Z0-9

/<короткий url> GET

В ответе - redirect (код ответа 302) на исходный URL

Хранение данных - в postgres. Схемы таблиц - на усмотрение разработчика.

В случае несуществующего коротокого URL - код 404.

### Задание 2

Реализовать сборку метрик через Prometheus.

Интересна частотность обоих запросов и время их обработки.

### Задание 3

Написать простой генератор нагрузки.
В виде последоватной Go-программы.

Нужно породить
* 10000 запросов на создание
* 100000 успешных GET-запросов (завершающихся редиректом)
* 100000 неудачных GET-запросов (завершающихся 404)


### Материалы

[Общее описание Gin Framework](https://gin-gonic.com/docs/)

[API Gin Framework](https://pkg.go.dev/github.com/gin-gonic/gin)

[Базы данных из Go](http://go-database-sql.org/)

[SQL API](https://pkg.go.dev/database/sql)

[Prometheus](https://prometheus.io/docs/introduction/overview/)

[Инструментирование Go-приложений для Prometheus](https://prometheus.io/docs/guides/go-application/)

[Пример простейшего http-клиента](https://gobyexample.com/http-clients)
