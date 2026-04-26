package api

import "net/http"

func Init() {
	// Публичные эндпоинты
	http.HandleFunc("/api/signin", signinHandler)      
	http.HandleFunc("/api/nextdate", NextDateHandler)  

	// Защищённые эндпоинты (через middleware Auth)
	http.HandleFunc("/api/task", Auth(taskHandler))           // POST/GET/PUT/DELETE
	http.HandleFunc("/api/tasks", Auth(tasksHandler))         // GET список
	http.HandleFunc("/api/task/done", Auth(doneTaskHandler))  // POST отметка
}