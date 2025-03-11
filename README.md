Необходимо написать сервис на Golang работающий по gRPC.
 Сервис должен:
 1) Принимать бинарные файлы (изображения) от клиента и сохранять их на жесткий
 диск.
 2) Иметь возможность просмотра списка всех загруженных файлов в формате:
 Имя файла | Дата создания | Дата обновления
 3) Отдавать файлы клиенту.
 4) Ограничивать количество одновременных подключений с клиента:- на загрузку/скачивание файлов - 10 конкурентных запросов;- на просмотр списка файлов - 100 конкурентных запросов.

 для запуска написать создать файл local.yaml, прописать туда необходимые параметры и запустить командой make run.