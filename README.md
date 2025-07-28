
## Запуск

1. Скачать зависимости
```bash
go mod tidy
```

2. Запустить проект командой
```bash
go run cmd/main.go
```

## Создание задачи
1. Создать задачу
```bash
curl -X POST localhost:8080/api/v1/task

# Ответ
# {"id":"2039bc69-ab6a-43c6-9697-048854202243"}

```

2. Используя полученный id привязать к задаче ссылки
```bash
 curl -H "Content-Type: application/json" -X POST localhost:8080/api/v1/task/2039bc69-ab6a-43c6-9697-048854202243/link 
 -d '
    [
        {
            "link":"https://upload.wikimedia.org/wikipedia/commons/4/41/Sunflower_from_Silesia2.jpg"
        },
        {
            "link":"https://upload.wikimedia.org/wikipedia/commons/thumb/3/3b/Europeian_diet_Sprite_bottle.jpg/800px-Europeian_diet_Sprite_bottle.jpg"
        },
        {
            "link":"https://edu.anarcho-copy.org/Programming%20Languages/Go/build-web-application-with-golang-en.pdf"
        }
    ]'

# Ответ
# {
#   "filesLink":[
#      {
#         "link":"https://upload.wikimedia.org/wikipedia/commons/4/41/Sunflower_from_Silesia2.jpg",
#         "status":"new"
#      },
#      {
#         "link":"https://upload.wikimedia.org/wikipedia/commons/thumb/3/3b/Europeian_diet_Sprite_bottle.jpg/800px-Europeian_diet_Sprite_bottle.jpg",
#         "status":"new"
#      },
#      {
#         "link":"https://edu.anarcho-copy.org/Programming%20Languages/Go/build-web-application-with-golang-en.pdf",
#         "status":"new"
#      }
#   ],
#   "id":"2039bc69-ab6a-43c6-9697-048854202243"
#}

``` 
3. Дождаться обработки задачи, статус можно проверить запросом
```bash
curl  localhost:8080/api/v1/task/2039bc69-ab6a-43c6-9697-048854202243 

# Ответ
# {
#   "filesLink":[
#      {
#         "link":"https://upload.wikimedia.org/wikipedia/commons/4/41/Sunflower_from_Silesia2.jpg",
#         "status":"completed"
#      },
#      {
#         "link":"https://upload.wikimedia.org/wikipedia/commons/thumb/3/3b/Europeian_diet_Sprite_bottle.jpg/800px-Europeian_diet_Sprite_bottle.jpg",
#         "status":"completed"
#      },
#      {
#         "link":"https://edu.anarcho-copy.org/Programming%20Languages/Go/build-web-application-with-golang-en.pdf",
#         "status":"completed"
#      }
#   ],
#   "id":"2039bc69-ab6a-43c6-9697-048854202243"
#}
```

3. Скчать архив с файлами с помощь id
```bash
curl --output ./arvhice.zip localhost:8080/api/v1/task/2039bc69-ab6a-43c6-9697-048854202243/result

```

## Особенности
1. Запуск обработки задач начинается только когда будет добавлено максимальное количество ссылок для задачи
2. Некоторые параметры настраиваются через переменные окружения, доступные настройки можно посмотреть в файле internal/config/config.go
3. Написал тест на проверку основного функционала
4. Фильтрацию контента реализовал через проверку ссылки, возможно стоит переделать что-бы тип файла определялся по Content-Type