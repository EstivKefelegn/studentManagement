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

func GetStudentsHndler(w http.ResponseWriter, r *http.Request) {

	var Students []models.Student
	Students, err := sqlconnect.GetStudentsDbHandler(Students, r)
	if err != nil {
		return
	}
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"Students"`
	}{
		Status: "success",
		Count:  len(Students),
		Data:   Students,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetStudentHndler(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	fmt.Println("ID is going to display")
	fmt.Println("The ID is: ", id)
	if err != nil {
		fmt.Println(err)
		return
	}

	Student, err := sqlconnect.GetStudentById(id)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Student)

}

func AddStudentsHandler(w http.ResponseWriter, r *http.Request) {

	var newStudents []models.Student
	var rawStudents []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error: reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &rawStudents)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Println(rawStudents)

	fields := GetFieldNames(models.Student{})
	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, Students := range rawStudents {
		for key := range Students {
			_, ok := allowedFields[key]
			if !ok {
				// http.Error(w, "Unacceptable field found in request. Only use allowed fields. ", http.StatusBadRequest)
				utils.ErrorHandler(err, "Unacceptable field found in request.")
				return
			}
		}
	}

	err = json.Unmarshal(body, &newStudents)

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON format: %v", err), http.StatusBadRequest)
		return
	}

	for _, student := range newStudents {
		err = CheckEmptyFields(student)
		if err != nil {
			return
		}
	}

	addTeacers, err := sqlconnect.AddStudentsDBHandler(newStudents)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string
		Count  int
		Data   []models.Student
	}{
		Status: "success",
		Count:  len(addTeacers),
		Data:   addTeacers,
	}

	json.NewEncoder(w).Encode(response)

}

// PUT func
func UpdateStudentHadler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		fmt.Println(w, "Invalid Student ID: ", http.StatusBadRequest)
		return
	}

	var updatedStudent models.Student
	err = json.NewDecoder(r.Body).Decode(&updatedStudent)
	if err != nil {
		// http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		utils.ErrorHandler(err, "Invalid request payload")
		return
	}

	updatedStudentDB, err := sqlconnect.UpdateStudentDBHandler(id, updatedStudent)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedStudentDB)
}

// patch Students

func PatchOneStudentsHandler(w http.ResponseWriter, r *http.Request) {
	// idStr := strings.TrimPrefix(r.URL.Path, "/Students/")
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid Student ID", http.StatusBadRequest)
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusInternalServerError)
	}

	updatedStudent, err := sqlconnect.PatchStudentDBHandler(id, updates)
	if err != nil {
		log.Println(err)
		return
	}

	json.NewEncoder(w).Encode(updatedStudent)

}

func PatchStudentsHandler(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "Invalid Student ID in update", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Could convert the id", http.StatusBadRequest)
			return
		}

		var StudentFormDB models.Student
		err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM Students WHERE id = ?", id).Scan(&StudentFormDB.ID,
			&StudentFormDB.FirstName, &StudentFormDB.LastName, &StudentFormDB.Email, &StudentFormDB.Class)

		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				http.Error(w, "Student not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error retriving Students", http.StatusInternalServerError)
			return
		}

		StudentsVal := reflect.ValueOf(&StudentFormDB).Elem()
		StudentType := StudentsVal.Type()

		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < StudentsVal.NumField(); i++ {
				field := StudentType.Field(i)
				// if field.Tag.Get("json") == k+",omitempty" {
				// if field.Tag.Get("json") == k {
				jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
				if jsonTag == k {
					fieldVal := StudentsVal.Field(i)
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
		fmt.Printf("Updating ID %d: %s, %s, %s, %s\n", StudentFormDB.ID,
			StudentFormDB.FirstName, StudentFormDB.LastName,
			StudentFormDB.Email, StudentFormDB.Class)

		res, err := tx.Exec("UPDATE Students SET first_name = ?, last_name = ?, email = ?, class = ? = ? WHERE id = ? ", StudentFormDB.FirstName,
			StudentFormDB.LastName, StudentFormDB.Email, StudentFormDB.Class, StudentFormDB.ID)

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

func DeleteStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/students/")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		// http.Error(w, "Request not found", http.StatusBadRequest)
		utils.ErrorHandler(err, "BadRequest")
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
		Status: "Student Successfully deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Println("Invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	deletedIDs, err := sqlconnect.DeleteStudents(ids)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status     string `json:"status"`
		DeletedIDs []int  `json:"deleted_IDs"`
	}{
		Status:     "Students successfully deleted",
		DeletedIDs: deletedIDs,
	}

	json.NewEncoder(w).Encode(response)

}

func StudentsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello get request from students page"))
		fmt.Println("hello get request from students page")
	case http.MethodPost:
		w.Write([]byte("Hello post request from student page"))
		fmt.Println("Hello post request from students page")
	case http.MethodPut:
		w.Write([]byte("Hello put request from student page"))
		fmt.Println("Hello put request from students page")
	case http.MethodDelete:
		w.Write([]byte("Hello delete request from student page"))
		fmt.Println("Hello delete requets from students page")
	}

}
