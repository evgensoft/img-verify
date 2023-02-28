[![Maintainability](https://api.codeclimate.com/v1/badges/390ad26a6778ff191ac5/maintainability)](https://codeclimate.com/github/evgensoft/img-verify/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/390ad26a6778ff191ac5/test_coverage)](https://codeclimate.com/github/evgensoft/img-verify/test_coverage)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=evgensoft_img-verify&metric=bugs)](https://sonarcloud.io/summary/new_code?id=evgensoft_img-verify)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=evgensoft_img-verify&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=evgensoft_img-verify)

Микросервис img-verify предназначен для предварительной проверки изображений загружаемых в соц. сети (большей частью в VK)

Сервис проверяет следующие параметры:
- [x] Проверка корректности загрузки изображения
- [x] Проверка размера изображения (длина+ширина не более 14000 пикселей, не более 50 Мб)
- [x] Проверка по хеш на изображения-заглушки Img-хостингов (Not_found, Img_deleted...)
- [x] Поиск лица человека на изображении
- [ ] Подключение сторонних face-detection
- [x] Проверка на качество изображения (https://cloud.yandex.ru/docs/vision/concepts)
