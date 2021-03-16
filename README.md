# HTTP Proxy Server

### Установка корневых сертификатов (Ubuntu)
`./scripts/certinstall.sh`

### Запуск

`docker build -t okto-proxy . && docker run -p 8080:8080 -p 8000:8000 okto-proxy`

### Проксирование HTTP и HTTPS запросов  
Прокси-сервер: 
`http://127.0.0.1:8080`
  
### Повторная отправка проксированных запросов  
* Список запросов: `http://127.0.0.1:8000/requests`
* Вывод запроса: `http://127.0.0.1:8000/requests/{id}`
* Повторная отправка запроса: `http://127.0.0.1:8000/repeat/{id}`
  
