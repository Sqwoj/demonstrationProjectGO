ИНСТРУКЦИЯ:


ЕСЛИ ЗАПУСКАЕМ ЧЕРЕЗ CMD/POWERSHELL:
ВПИСЫВАЕМ ПЕРЕМЕННЫЕ ОКРУЖЕНИЯ
$env:JWT_SECRET_KEY="super-secret-key-1234567890"
$env:DB_HOST='127.0.0.1'
$env:DB_PORT='5432'
$env:DB_USER='postgres'
$env:DB_PASSWORD='1842'
$env:DB_NAME='postgres'

ЗАПУСК ЧЕРЕЗ start.bat ВПИСЫВАЕТ ПЕРЕМЕННЫЕ ОКРУЖЕНИЯ САМ, ПОЭТОМУ ПРОСТО ЗАПУСКАЕМ

СОБИРАЕМ ПРОЕКТ
go build .

ЗАПУСКАЕМ ЕГО
go run .

ПРОВЕРЯЕМ РЕГИСТРАЦИЯ
$regBody = @{ username = "alice"; password = "secret123" } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/register" -ContentType "application/json" -Body $regBody

ПРОВЕРКА ЛОГИНА
$loginBody = @{ username = "alice"; password = "secret123" } | ConvertTo-Json
$login = Invoke-RestMethod -Method Post -Uri "http://localhost:8080/login" -ContentType "application/json" -Body $loginBody
$token = $login.token

ПОЛУЧАЕТ СПИСОК УСТРОЙСТВ
Invoke-RestMethod -Method Get -Uri "http://localhost:8080/devices" -Headers @{ Authorization = "Bearer $token" }

СПИСОК ИВЕНТОВ ПО ID УСТРОЙСТВА
Invoke-RestMethod -Method Get -Uri "http://localhost:8080/events?device_id=$($device.id)" -Headers @{ Authorization = "Bearer $token" }

ПРОВЕРКА НЕВЕРНОГО ТОКЕНА
Invoke-RestMethod -Method Get -Uri "http://localhost:8080/devices" -Headers @{ Authorization = "Bearer fake-token" }

СОЗДАЕМ УСТРОЙСТВО
$deviceBody = @{ name = "Core-SW-01"; ip = "10.0.0.10"; status = "online" } | ConvertTo-Json
$device = Invoke-RestMethod -Method Post -Uri "http://localhost:8080/devices" -Headers @{ Authorization = "Bearer $token" } -ContentType "application/json" -Body $deviceBody

СОЗДАЕМ ИВЕНТ
$eventBody = @{ device_id = $device.id; description = "Link up" } | ConvertTo-Json
$event = Invoke-RestMethod -Method Post -Uri "http://localhost:8080/events" -Headers @{ Authorization = "Bearer $token" } -ContentType "application/json" -Body $eventBody
