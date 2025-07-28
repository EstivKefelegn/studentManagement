package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"student_management_api/Golang/internal/models"
	"student_management_api/Golang/internal/repository/sqlconnect"
)

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"calss":      true,
		"subject":    true,
	}

	return validFields[field]
}
func GetTeachersHndler(w http.ResponseWriter, r *http.Request) {

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Couldn't connect to db: %v", err), http.StatusBadRequest)
	}
	defer db.Close()

	query := `SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1`
	var args []interface{}

	query, args = addFilter(r, query, args)

	query = addSorting(r, query)
	rows, err := db.Query(query, args...)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Databse Query Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	teachersList := make([]models.Teacher, 0)

	for rows.Next() {
		var teacher models.Teacher
		err = rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			http.Error(w, "Error scanning databse results", http.StatusInternalServerError)
			return
		}
		teachersList = append(teachersList, teacher)
		fmt.Println(teachersList)
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

}

func GetTeacherHndler(w http.ResponseWriter, r *http.Request) {

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Couldn't connect to db: %v", err), http.StatusBadRequest)
	}
	defer db.Close()

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	fmt.Println("ID is going to display")
	fmt.Println("The ID is: ", id)
	if err != nil {
		fmt.Println(err)
		return
	}

	var teacher models.Teacher
	err = db.QueryRow(`SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?`, id).Scan(
		&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)

	if err == sql.ErrNoRows {
		http.Error(w, "No rows found with this ID", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Databse query error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)

}

func addSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 { // if sortvalue exists
		query += " ORDER BY"
		for i, param := range sortParams {
			// /?sortby=last_name:asc
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortField(field) || !isValidSortOrder(order) {
				continue
			}

			if i > 0 {
				query += ","
			}
			query += " " + field + " " + order
		}
	}
	return query
}

func addFilter(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"calss":      "class",
		"subject":    "subject",
	}

	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)
		}
	}

	return query, args
}

func AddTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()

	if err != nil {
		http.Error(w, "Couldn't connect to the database", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	var newTeachers []models.Teacher
	err = json.NewDecoder(r.Body).Decode(&newTeachers)

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON format: %v", err), http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare(`INSERT INTO teachers(first_name, last_name, email, class, subject) VALUES(?,?,?,?,?)`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database preparation failed: %v", err), http.StatusBadRequest)
		return
	}

	defer stmt.Close()

	addTeacers := make([]models.Teacher, len(newTeachers))

	for i, teacher := range newTeachers {
		res, err := stmt.Exec(teacher.FirstName, teacher.LastName, teacher.Email, teacher.Class, teacher.Subject)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, "Couldnt fetch the ID", http.StatusInternalServerError)
		}
		teacher.ID = int(id)
		addTeacers[i] = teacher

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
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
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

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Println("Couldn't connect to the database")
		http.Error(w, "Unable to connect", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	var existingTeachers models.Teacher
	row := db.QueryRow(`SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?`, id)

	err = row.Scan(&existingTeachers.ID, &existingTeachers.FirstName, &existingTeachers.LastName, &existingTeachers.Email, &existingTeachers.Class, &existingTeachers.Subject)

	if err == sql.ErrNoRows {
		http.Error(w, "No row's found with this id", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Unable to retrive the data", http.StatusInternalServerError)
		return
	}
	updatedTeacher.ID = existingTeachers.ID
	res, err := db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, updatedTeacher.ID)

	if err != nil {
		http.Error(w, "Uable to update the db", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Rows updated: ", rowsAffected)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacher)
}

// patch teachers

func PatchOneTeachersHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusInternalServerError)
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Couldn't connect to thw database", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	var existingTeachers models.Teacher
	row := db.QueryRow(`SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?`, id)
	err = row.Scan(&existingTeachers.ID, &existingTeachers.FirstName, &existingTeachers.LastName, &existingTeachers.Email, &existingTeachers.Class, &existingTeachers.Subject)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No rows foun with this ID", http.StatusNotFound)
			return
		}

		http.Error(w, "Unable to retrive data", http.StatusInternalServerError)
		return
	}

	// for k, v := range updates {
	// 	switch k {
	// 	case "first_name":
	// 		existingTeachers.FirstName = v.(string)
	// 	case "last_name":
	// 		existingTeachers.LastName = v.(string)
	// 	case "email":
	// 		existingTeachers.Email = v.(string)
	// 	case "subject":
	// 		existingTeachers.Subject = v.(string)

	// 	}
	// }

	teacherVal := reflect.ValueOf(&existingTeachers).Elem()
	teacherType := teacherVal.Type()

	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherType.Field(i)
			fmt.Println("The current field is: ", field)

			if field.Tag.Get("json") == k+",omitempty" {
				if teacherVal.Field(i).CanSet() {
					fieldVal := teacherVal.Field(i)
					fieldVal.Set(reflect.ValueOf(v).Convert(teacherVal.Field(i).Type()))
				}
			}

		}
	}
	fmt.Println(teacherVal.Type())

	res, err := db.Exec(`UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?`, existingTeachers.FirstName, existingTeachers.LastName, existingTeachers.Email, existingTeachers.Class, existingTeachers.Subject, existingTeachers.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Couldn't update the reuested data: %v", err), http.StatusNotFound)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Rows updated: ", rowsAffected)

	json.NewEncoder(w).Encode(updates)

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

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Couldn't connect to db", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	res, err := db.Exec(`DELETE FROM teachers WHERE id = ?`, id)

	if err != nil {
		http.Error(w, "No request data found", http.StatusNotFound)
		return
	}

	affectedRow, err := res.RowsAffected()
	if err != nil {
		fmt.Println("No affected row is found")
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
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Couldn't connect to db", http.StatusInternalServerError)
		return
	}

	defer db.Close()
	var ids []int
	err = json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Println("Invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest) 
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println("Invlid transaction")
		http.Error(w, "Invalid transaction", http.StatusInternalServerError)
		return
	}

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid transaction", http.StatusInternalServerError)
		return
	}

	defer stmt.Close()

	deletedIDs := []int{}

	for _, id := range ids {
		res, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			log.Println("Invlid transaction")
			http.Error(w, "Invalid transaction", http.StatusInternalServerError)
			return
		}

		affectedRow, err := res.RowsAffected()
		if err != nil {
			log.Println("Invlid transaction")
			http.Error(w, "Invalid transaction", http.StatusInternalServerError)
			return
		}

		if affectedRow > 0 {
			deletedIDs = append(deletedIDs, id)
		}

		if affectedRow < 1 {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("ID %d doesn't exist", id), http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("err")
		http.Error(w, "Error commiting transaction", http.StatusInternalServerError)
		return
	}

	if len(deletedIDs) < 1 {
		http.Error(w, "IDs do not exists", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status     string `json:"status"`
		DeletedIDs []int `json:"deleted_IDs`
	}{
		Status: "Teachers successfully deleted",
		DeletedIDs: deletedIDs,
	}

	json.NewEncoder(w).Encode(response)

}

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// w.Write([]byte("Hello get request from teachers page"))
		GetTeacherHndler(w, r)
	case http.MethodPost:
		// w.Write([]byte("Hello post request from teachers page"))
		AddTeachersHandler(w, r)
	case http.MethodPut:
		UpdateTeacherHadler(w, r)
		// w.Write([]byte("Hello put request from teachers page"))
		// fmt.Println("Hello put request from teachers page")
	case http.MethodPatch:
		PatchOneTeachersHandler(w, r)
	case http.MethodDelete:
		DeleteTeachersHandler(w, r)
	}
}
