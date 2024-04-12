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
	el comando para correrlo es <--CompileDaemon-->*/
	"github.com/tealeg/xlsx" //con este import es con el cual podemos hacer la manipulacion de nuestros Excels este es el comando para instalarlo <--go get -u github.com/tealeg/xlsx-->
)

// esto es una tarea aqui es donde deberiamos hacer la consulta en vez de hardcodear uno jeje
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

// Funcion para obtener los datos de  las tareas desde el excel
func readTasksFromExcel(filePath string) ([]task, error) {
	// Abre el archivo Excel
	xlFile, err := xlsx.OpenFile("DB_Excel.xlsx")
	if err != nil {
		return nil, err
	}

	var tasks []task

	// Itera sobre las hojas del archivo Excel
	for _, sheet := range xlFile.Sheets {
		// Itera sobre las filas de la hoja
		for rowIndex, row := range sheet.Rows {
			if rowIndex == 0 {
				// Ignora la primera fila que generalmente contiene los encabezados
				continue
			}

			var newTask task
			// Itera sobre las celdas de la fila
			for cellIndex, cell := range row.Cells {
				switch cellIndex {
				case 0:
					newTask.ID, _ = cell.Int()
				case 1:
					newTask.Name = cell.String()
				case 2:
					newTask.Content = cell.String()
				}
			}
			// Agrega la tarea a la lista
			tasks = append(tasks, newTask)
		}
	}
	return tasks, nil
}

// esta es la funcion que se encargara de mostrar las tareas

// Función para obtener todas las tareas desde el archivo Excel
func getTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := readTasksFromExcel("DB_Excel.xlsx") // Lee las tareas desde el archivo Excel
	if err != nil {
		http.Error(w, "Error al leer las tareas del archivo Excel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json") // Establece el tipo de contenido en la cabecera de la respuesta
	json.NewEncoder(w).Encode(tasks)                   // Codifica las tareas en formato JSON y las envía como respuesta
}

// Funcion Para crear Tareas
func createTask(w http.ResponseWriter, r *http.Request) {
	var newTask task

	// Lee el cuerpo de la solicitud
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error al leer el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	// Decodifica los datos de la solicitud en una estructura de tarea
	err = json.Unmarshal(reqBody, &newTask)
	if err != nil {
		http.Error(w, "Error al decodificar los datos de la tarea", http.StatusBadRequest)
		return
	}

	// Abre el archivo Excel
	xlFile, err := xlsx.OpenFile("DB_Excel.xlsx")
	if err != nil {
		http.Error(w, "Error al abrir el archivo Excel", http.StatusInternalServerError)
		return
	}

	// Obtiene la hoja de Excel
	var sheet *xlsx.Sheet
	if len(xlFile.Sheets) == 0 {
		// Si no hay hojas, crea una nueva hoja llamada "Tarea"
		sheet, err = xlFile.AddSheet("Tarea")
		if err != nil {
			http.Error(w, "Error al agregar una nueva hoja al archivo Excel", http.StatusInternalServerError)
			return
		}
	} else {
		// Si ya hay una hoja, utiliza la primera hoja encontrada
		sheet = xlFile.Sheets[0]
	}

	// Encuentra el ID más alto en la hoja de Excel
	highestID := 0
	for _, row := range sheet.Rows {
		if len(row.Cells) > 0 {
			id, _ := row.Cells[0].Int()
			if id > highestID {
				highestID = id
			}
		}
	}

	// Asigna el nuevo ID para la tarea
	newTask.ID = highestID + 1

	// Crea una nueva fila para la nueva tarea
	row := sheet.AddRow()

	// Agrega las celdas para el nuevo registro
	cell := row.AddCell()
	cell.SetInt(newTask.ID)

	cell = row.AddCell()
	cell.SetString(newTask.Name)

	cell = row.AddCell()
	cell.SetString(newTask.Content)

	// Guarda los cambios en el archivo Excel
	err = xlFile.Save("DB_Excel.xlsx")
	if err != nil {
		http.Error(w, "Error al guardar los cambios en el archivo Excel", http.StatusInternalServerError)
		return
	}

	// Responde con la tarea creada
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

// Funcion para traer tareas con su ID
func getTaskByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    taskID, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }

    // Abre el archivo Excel
    xlFile, err := xlsx.OpenFile("DB_Excel.xlsx")
    if err != nil {
        http.Error(w, "Error opening Excel file", http.StatusInternalServerError)
        return
    }

    // Obtiene la hoja de tareas
    sheet, err := getSheetByName(xlFile, "Tarea")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var foundTask task
    found := false

    // Itera sobre las filas para encontrar la tarea con el ID dado
    for _, row := range sheet.Rows {
        idCell := row.Cells[0]
        id, _ := idCell.Int()
        if id == taskID {
            foundTask.ID = id
            foundTask.Name = row.Cells[1].String()
            foundTask.Content = row.Cells[2].String()
            found = true
            break
        }
    }

    // Si no se encontró la tarea, devuelve un error
    if !found {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }

    // Retorna la tarea encontrada
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(foundTask)
}


// Funcion para eliminar tareas segun su ID
// Luego, en tu función deleteTask, puedes usar esta función así:
func deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID de tarea inválido", http.StatusBadRequest)
		return
	}

	// Abre el archivo Excel
	xlFile, err := xlsx.OpenFile("DB_Excel.xlsx")
	if err != nil {
		http.Error(w, "Error al abrir el archivo Excel", http.StatusInternalServerError)
		return
	}

	// Obtiene la hoja de tareas
	sheet, err := getSheetByName(xlFile, "Tarea")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Busca la fila correspondiente al ID de la tarea y la elimina
	found := false
	for i, row := range sheet.Rows {
		if i == 0 {
			// Ignora la primera fila de encabezados
			continue
		}
		idCell := row.Cells[0]
		id, _ := idCell.Int()
		if id == taskID {
			sheet.RemoveRowAtIndex(i)
			found = true
			break
		}
	}

	// Si no se encontró la tarea con el ID dado, devuelve un error
	if !found {
		http.Error(w, "Tarea no encontrada", http.StatusNotFound)
		return
	}

	// Guarda los cambios en el archivo Excel
	err = xlFile.Save("DB_Excel.xlsx")
	if err != nil {
		http.Error(w, "Error al guardar los cambios en el archivo Excel", http.StatusInternalServerError)
		return
	}

	// Respuesta exitosa
	w.WriteHeader(http.StatusOK)
}

// Funcion para Actualizar tareas segun su ID
func updateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var newTask task // Use the correct struct type
	err = json.NewDecoder(r.Body).Decode(&newTask)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	xlFile, err := xlsx.OpenFile("DB_Excel.xlsx")
	if err != nil {
		http.Error(w, "Error opening Excel file", http.StatusInternalServerError)
		return
	}

	sheet, err := getSheetByName(xlFile, "Tarea")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	found := false
	for _, row := range sheet.Rows {
		idCell := row.Cells[0]
		id, _ := idCell.Int()
		if id == taskID {
			row.Cells[1].Value = newTask.Name
			row.Cells[2].Value = newTask.Content // Assuming Content is the field name
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	err = xlFile.Save("DB_Excel.xlsx")
	if err != nil {
		http.Error(w, "Error saving Excel file", http.StatusInternalServerError)
		return
	}
}

// Esta es la pagina que habre cuando se realiza una peticion en la pagina Raiz
func homeLink(w http.ResponseWriter, r *http.Request) { //la w es la respuesta que podremos darle al cliente y un objeto R que sera la peticion del usuario por eso sus tipos
	fmt.Fprintf(w, "Welcome to my API! 3DOR33")
}

// Función para buscar una hoja de cálculo por su nombre
func getSheetByName(file *xlsx.File, name string) (*xlsx.Sheet, error) {
	for _, sheet := range file.Sheets {
		if sheet.Name == name {
			return sheet, nil
		}
	}
	return nil, fmt.Errorf("Hoja de cálculo '%s' no encontrada", name)
}

func main() { //Es necesario especificar el Methods que usara cada ruta
	router := mux.NewRouter().StrictSlash(true)                    //aqui ponemos el routeador en modo estricto para que las rutas las lea como las tenemos declaradas
	router.HandleFunc("/", homeLink)                               //aqui le decimos que la ruta principal sera la funcion homeLink (este es el enrouteador que creamos)
	router.HandleFunc("/tasks", getTasks).Methods("GET")           //aqui le decimos que la ruta /tasks sera la funcion getTasks y que sera de tipo GET
	router.HandleFunc("/tasks", createTask).Methods("POST")        //aqui le decimos que la ruta /tasks sera la funcion createTask y que sera de tipo POST
	router.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE") //aqui le decimos que la ruta /tasks sera la funcion deleteTask y que sera de tipo DELETE
	router.HandleFunc("/tasks/{id}", getTaskByID).Methods("GET")   //aqui le decimos que la ruta /tasks sera la funcion createTask y que sera de tipo POST
	router.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")    //aqui le decimos que la ruta /tasks sera la funcion updateTask y que sera de tipo PUT
	http.ListenAndServe(":3303", router)                           //aqui le decimos que escuche en el puerto 3000 y que use el routeador que creamos (router)
	log.Fatal(http.ListenAndServe(":3303", router))                //aqui le decimos que si hay un error que lo muestre en consola
}
