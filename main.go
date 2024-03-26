package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"      //aqui importamos el paquete log para poder hacer uso de los metodos de log
	"net/http" //aqui importamos el paquete http para poder hacer uso de los metodos de http
	"strconv"  //aqui importamos el paquete strconv para poder convertir un string a un entero

	"github.com/gorilla/mux" //aqui importamos el mux para tener el routeador este es comando <--go get -u github.com/gorilla/mux-->
	/*este es solamente un comentario para decir que intalamos go get github.com/githubnemo/CompileDaemon
	para que se reinicie el servidor cada vez que guardemos un cambio vendria siendo como un nodemon en nodejs
	el comando para correrlo es <--CompileDaemon-->*/)

// esto e una tarea aqui es donde deveriamos hacer la construcion a nuestra base de datos
type task struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// Aqui se crea un slice de tareas
type allTasks []task

var tasks = allTasks{
	{
		ID:      1,
		Name:    "Task One",
		Content: "Some Content",
	},
}

// esta es la funcion que se encargara de mostrar las tareas
func getTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") //aqui mandamos el tipo de contenido que se mostrara en la cabecera de la respuesta
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var newTask task
	reqBody, err := ioutil.ReadAll(r.Body) //aqui leemos la peticion del usuario y vemos si hay algun error o si no hay error guardamos el reqBody
	if err != nil {
		fmt.Fprintf(w, "Insert a Valid Task") //aqui le decimos que si hay un error que muestre este mensaje
	}
	json.Unmarshal(reqBody, &newTask)                  //aqui le decimos que tome el reqBody y lo guarde en la variable newTask
	newTask.ID = len(tasks) + 1                        //aqui le decimos que el ID de la nueva tarea sera la longitud de las tareas mas 1
	tasks = append(tasks, newTask)                     //aqui le decimos que agregue la nueva tarea al slice de tareas
	w.Header().Set("Content-Type", "application/json") //aqui mandamos el tipo de contenido que se mostrara en la cabecera de la respuesta
	w.WriteHeader(http.StatusCreated)                  //aqui le decimos que muestre el estado de la creacion de la tarea
	json.NewEncoder(w).Encode(newTask)                 //aqui le decimos que muestre la nueva tarea que se creo
}
func getTaskbyID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                     //esta es la variable que se encargara de obtener el id de la tarea de la ruta
	taskID, err := strconv.Atoi(vars["id"]) //aqui le decimos que convierta el id de la tarea a un entero

	if err != nil {
		fmt.Fprintf(w, "Invalid Task ID") //aqui le decimos que si hay un error que muestre este mensaje
		return
	}
	//aqui le decimos que recorra todas las tareas
	for _, task := range tasks {
		//aqui le decimos que si el ID de la tarea es igual al ID de la tarea que se busca
		if task.ID == taskID {
			w.Header().Set("Content-Type", "application/json") //aqui mandamos el tipo de contenido que se mostrara en la cabecera de la respuesta
			json.NewEncoder(w).Encode(task)                    //aqui le decimos que muestre la tarea que se busca
		}

	}
}
func deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                     //esta es la variable que se encargara de obtener el id de la tarea de la ruta
	taskID, err := strconv.Atoi(vars["id"]) //aqui le decimos que convierta el id de la tarea a un entero

	if err != nil {
		fmt.Fprintf(w, "Invalid Task ID") //aqui le decimos que si hay un error que muestre este mensaje
		return
	}
	//aqui le decimos que recorra todas las tareas
	for i, task := range tasks {
		//aqui le decimos que si el ID de la tarea es igual al ID de la tarea que se busca
		if task.ID == taskID {
			tasks = append(tasks[:i], tasks[i+1:]...)                  //aqui le decimos que borre la tarea que se busca
			fmt.Fprintf(w, "Se elimino la task con el id: %v", taskID) //aqui le decimos que muestre el estado de la eliminacion de la tarea
		}

	}
}
func updateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var updatedTask task

	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Fprintf(w, "Invalid Task ID")
		return
	}
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Please enter Valid Data")
		return
	}
	json.Unmarshal(reqBody, &updatedTask)

	for i, t := range tasks {
		if t.ID == taskID {
			tasks = append(tasks[:i], tasks[i+1:]...)
			updatedTask.ID = taskID
			tasks = append(tasks, updatedTask)

			fmt.Fprintf(w, "The task with ID %v has been updated successfully", taskID)
		}
	}

}

func homeLink(w http.ResponseWriter, r *http.Request) { //la w es la respuesta que podremos darle al cliente y un objeto R que sera la peticion del usuario por eso sus tipos
	fmt.Fprintf(w, "Welcome to my API! 3DOR33")

}

func main() { //Es necesario especificar el Methods que usara cada ruta
	router := mux.NewRouter().StrictSlash(true)             //aqui ponemos el routeador en modo estricto para que las rutas las lea como las tenemos declaradas
	router.HandleFunc("/", homeLink)                        //aqui le decimos que la ruta principal sera la funcion homeLink (este es el enrouteador que creamos)
	router.HandleFunc("/tasks", getTasks).Methods("GET")    //aqui le decimos que la ruta /tasks sera la funcion getTasks y que sera de tipo GET
	router.HandleFunc("/tasks", createTask).Methods("POST") //aqui le decimos que la ruta /tasks sera la funcion createTask y que sera de tipo POST
	router.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE") //aqui le decimos que la ruta /tasks sera la funcion deleteTask y que sera de tipo DELETE
	router.HandleFunc("/tasks/{id}", getTaskbyID).Methods("GET") //aqui le decimos que la ruta /tasks sera la funcion createTask y que sera de tipo POST
	router.HandleFunc("/tasks/{id}", updateTask).Methods("PUT") //aqui le decimos que la ruta /tasks sera la funcion updateTask y que sera de tipo PUT
	http.ListenAndServe(":3303", router)                         //aqui le decimos que escuche en el puerto 3000 y que use el routeador que creamos (router)
	log.Fatal(http.ListenAndServe(":3303", router))              //aqui le decimos que si hay un error que lo muestre en consola
}
