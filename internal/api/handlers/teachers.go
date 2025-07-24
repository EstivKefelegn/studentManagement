package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"student_management_api/Golang/internal/models"
	"sync"
)

var (
	teachers = make(map[int]models.Teacher)
	mutex    = &sync.Mutex{}
	nextID   = 1
)

func init() {
	teachers[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "john",
		LastName:  "Doe",
		Class:     "10A",
		Subject:   "Algebra",
	}
	nextID++
	teachers[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "Estiv",
		LastName:  "Kefelegn",
		Class:     "15A",
		Subject:   "Eng",
	}
	nextID++
	teachers[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "jahn",
		LastName:  "Doe",
		Class:     "10A",
		Subject:   "Math",
	}
	nextID++
}

func getTeacherHndler(w http.ResponseWriter, r *http.Request) {

	first_name := r.URL.Query().Get("first_name")
	last_name := r.URL.Query().Get("last_name")
	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	fmt.Println(idStr)
	fmt.Println("Firts Name", first_name)
	fmt.Println("Last Name", last_name)
	teachersList := make([]models.Teacher, 0, len(teachers))
	if idStr == "" {
		for _, teacher := range teachers {
			if (first_name == "" || teacher.FirstName == first_name) && (last_name == "" || teacher.LastName == last_name) {
				teachersList = append(teachersList, teacher)
			}
		}
		response := struct {
			Status string           `json:"status"`
			Count  int              `json:"count"`
			Data   []models.Teacher `json:"teachers"`
		}{
			Status: "success",
			Count:  len(teachersList),
			Data:   teachersList,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	id, err := strconv.Atoi(idStr)
	fmt.Println("ID is going to display")
	fmt.Println("The ID is: ", id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("The ID is: ", id)
	teacher, exists := teachers[id]
	if !exists {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return 
	}

	json.NewEncoder(w).Encode(teacher)

}

func addTeachersHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var newTeachers []models.Teacher
	err := json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	addedTeachers := make([]models.Teacher, len(newTeachers))

	for i, newTeacher := range newTeachers {
		newTeacher.ID = nextID
		teachers[nextID] = newTeacher
		addedTeachers[i] = newTeacher
		nextID++
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "sucess",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}

	json.NewEncoder(w).Encode(response)
}


func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// w.Write([]byte("Hello get request from teachers page"))
		getTeacherHndler(w, r)
	case http.MethodPost:
		// w.Write([]byte("Hello post request from teachers page"))
		addTeachersHandler(w, r)
	case http.MethodPut:
		w.Write([]byte("Hello put request from teachers page"))
		fmt.Println("Hello put request from teachers page")
	case http.MethodDelete:
		w.Write([]byte("Hello delete request from teachers page"))
		fmt.Println("Hello delete requets from teachers page")
	}
}