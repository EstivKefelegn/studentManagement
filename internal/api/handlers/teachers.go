package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"student_management_api/Golang/internal/models"
	"student_management_api/Golang/internal/repository/sqlconnect"
	"student_management_api/Golang/pkg/utils"
)

func GetTeachersHndler(w http.ResponseWriter, r *http.Request) {

	var teachers []models.Teacher
	teachers, err := sqlconnect.GetTeachersDbHandler(teachers, r)
	if err != nil {
		return
	}
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"teachers"`
	}{
		Status: "success",
		Count:  len(teachers),
		Data:   teachers,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetTeacherHndler(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	fmt.Println("ID is going to display")
	fmt.Println("The ID is: ", id)
	if err != nil {
		fmt.Println(err)
		return
	}

	teacher, err := sqlconnect.GetTeacherById(id)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)

}

func AddTeachersHandler(w http.ResponseWriter, r *http.Request) {

	var newTeachers []models.Teacher
	var rawTeachers []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error: reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &rawTeachers)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Println(rawTeachers)

	fields := GetFieldNames(models.Teacher{})

	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, teachers := range rawTeachers {
		for key := range teachers {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable field found in request. Only use allowed fields. ", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newTeachers)

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON format: %v", err), http.StatusBadRequest)
		return
	}

	for _, teacher := range newTeachers {
		err = CheckEmptyFields(teacher)
		if err != nil {
			return
		}
	}

	addTeacers, err := sqlconnect.AddTeachersDBHandler(newTeachers)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string
		Count  int
		Data   []models.Teacher
	}{
		Status: "success",
		Count:  len(addTeacers),
		Data:   addTeacers,
	}

	json.NewEncoder(w).Encode(response)

}

// PUT func
func UpdateTeacherHadler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		fmt.Println(w, "Invalid Teacher ID: ", http.StatusBadRequest)
		return
	}

	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	updatedTeacherDB, err := sqlconnect.UpdateTeacherDBHandler(id, updatedTeacher)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacherDB)
}

// patch teachers

func PatchOneTeachersHandler(w http.ResponseWriter, r *http.Request) {
	// idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusInternalServerError)
	}

	updatedTeacher, err := sqlconnect.PatchOneStudentDBHandler(id, updates)
	if err != nil {
		log.Println(err)
		return
	}

	json.NewEncoder(w).Encode(updatedTeacher)

}

func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	var updates []map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error starting transactions", http.StatusInternalServerError)
		return
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			http.Error(w, "Invalid teacher ID in update", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Could convert the id", http.StatusBadRequest)
			return
		}

		var teacherFormDB models.Teacher
		err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(&teacherFormDB.ID,
			&teacherFormDB.FirstName, &teacherFormDB.LastName, &teacherFormDB.Email, &teacherFormDB.Class, &teacherFormDB.Subject)

		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				http.Error(w, "Teacher not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error retriving teachers", http.StatusInternalServerError)
			return
		}

		teachersVal := reflect.ValueOf(&teacherFormDB).Elem()
		teacherType := teachersVal.Type()

		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < teachersVal.NumField(); i++ {
				field := teacherType.Field(i)
				// if field.Tag.Get("json") == k+",omitempty" {
				// if field.Tag.Get("json") == k {
				jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
				if jsonTag == k {
					fieldVal := teachersVal.Field(i)
					fmt.Println("The field value is: ", fieldVal)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(v)
						if val.Type().ConvertibleTo(field.Type) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							log.Printf("cannot convert %v to %v", val.Type(), fieldVal.Type())
							return
						}
					}
					break
				}
			}
		}
		fmt.Printf("Updating ID %d: %s, %s, %s, %s, %s\n", teacherFormDB.ID,
			teacherFormDB.FirstName, teacherFormDB.LastName,
			teacherFormDB.Email, teacherFormDB.Class, teacherFormDB.Subject)

		res, err := tx.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ? ", teacherFormDB.FirstName,
			teacherFormDB.LastName, teacherFormDB.Email, teacherFormDB.Class, teacherFormDB.Subject, teacherFormDB.ID)

		if err != nil {
			tx.Rollback()
			http.Error(w, "Couldn't update the value", http.StatusInternalServerError)
			return
		}

		affectedRows, err := res.RowsAffected()
		if err != nil {
			fmt.Println("There is no affectedRows")
			return
		}

		fmt.Println(affectedRows, "Affected Rows")

	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Error commiting transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Request not found", http.StatusBadRequest)
		return
	}

	affectedRow, err := sqlconnect.DeleteOneStudent(w, id)
	if err != nil {
		return
	}

	fmt.Println(affectedRow, "Affected row")
	// w.WriteHeader(http.StatusNoContent)

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Teacher Successfully deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Println("Invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	deletedIDs, err := sqlconnect.DeleteStudents(ids)
	if err != nil {
		utils.ErrorHandler(err, "Can't delete a teacher")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status     string `json:"status"`
		DeletedIDs []int  `json:"deleted_IDs"`
	}{
		Status:     "Teachers successfully deleted",
		DeletedIDs: deletedIDs,
	}

	json.NewEncoder(w).Encode(response)

}

func GetStudentsByTeacherID(w http.ResponseWriter, r *http.Request) {
	teacherID := r.PathValue("id")
	var students []models.Student

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Println(err)
		return
	}

	defer db.Close()

	students, err = sqlconnect.GetStudentsByTeacherIDFromDB(teacherID, students)
	if err != nil {
		return
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "Success",
		Count:  len(students),
		Data:   students,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetStudentCountByTeacherID(w http.ResponseWriter, r *http.Request) {
	teacherID := r.PathValue("id")
	count, err := sqlconnect.GetStudentCountByTeachersIDFromDB(teacherID)
	if err != nil {
		log.Println(err)
		return
	}

	response := struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}{
		Status: "success",
		Count:  count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetTeacherHndler(w, r)
	case http.MethodPost:
		AddTeachersHandler(w, r)
	case http.MethodPut:
		UpdateTeacherHadler(w, r)
	case http.MethodPatch:
		PatchOneTeachersHandler(w, r)
	case http.MethodDelete:
		DeleteTeachersHandler(w, r)
	}
}
