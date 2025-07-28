package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"student_management_api/Golang/internal/models"
)

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

func GetTeachersDbHandler(teachers []models.Teacher, r *http.Request) ([]models.Teacher, error) {
	db, err := ConnectDB()

	if err != nil {
		// http.Error(w, fmt.Sprintf("Couldn't connect to db: %v", err), http.StatusBadRequest)
		return nil, err
	}
	defer db.Close()

	query := `SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1`
	var args []interface{}

	query, args = addFilter(r, query, args)

	query = addSorting(r, query)

	rows, err := db.Query(query, args...)

	if err != nil {
		fmt.Println(err)
		// http.Error(w, "Databse Query Error", http.StatusInternalServerError)
		return nil, err
	}
	defer rows.Close()

	teachersList := make([]models.Teacher, 0)

	for rows.Next() {
		var teacher models.Teacher
		err = rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			// http.Error(w, "Error scanning databse results", http.StatusInternalServerError)
			return nil, err
		}
		teachersList = append(teachersList, teacher)
		fmt.Println(teachersList)
	}
	return teachersList, nil
}

func GetTeacherById(w http.ResponseWriter, id int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Couldn't connect to db: %v", err), http.StatusBadRequest)
	}
	defer db.Close()

	var teacher models.Teacher
	err = db.QueryRow(`SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?`, id).Scan(
		&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)

	if err == sql.ErrNoRows {
		http.Error(w, "No rows found with this ID", http.StatusNotFound)
		return models.Teacher{}, err
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Databse query error: %v", err), http.StatusInternalServerError)
		return models.Teacher{}, err
	}
	return teacher, nil
}

func AddTeachersDBHandler(w http.ResponseWriter, newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()

	if err != nil {
		http.Error(w, "Couldn't connect to the database", http.StatusInternalServerError)
		return nil, err
	}

	defer db.Close()

	stmt, err := db.Prepare(`INSERT INTO teachers(first_name, last_name, email, class, subject) VALUES(?,?,?,?,?)`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database preparation failed: %v", err), http.StatusBadRequest)
		return nil, err
	}

	defer stmt.Close()

	addTeacers := make([]models.Teacher, len(newTeachers))

	for i, teacher := range newTeachers {
		res, err := stmt.Exec(teacher.FirstName, teacher.LastName, teacher.Email, teacher.Class, teacher.Subject)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, "Couldnt fetch the ID", http.StatusInternalServerError)
		}
		teacher.ID = int(id)
		addTeacers[i] = teacher

	}
	return addTeacers, err
}

func UpdateTeacherDBHandler(w http.ResponseWriter, id int, updatedTeacher models.Teacher) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Couldn't connect to the database")
		http.Error(w, "Unable to connect", http.StatusInternalServerError)
		return models.Teacher{}, err
	}

	defer db.Close()

	var existingTeachers models.Teacher
	row := db.QueryRow(`SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?`, id)

	err = row.Scan(&existingTeachers.ID, &existingTeachers.FirstName, &existingTeachers.LastName, &existingTeachers.Email, &existingTeachers.Class, &existingTeachers.Subject)

	if err == sql.ErrNoRows {
		http.Error(w, "No row's found with this id", http.StatusBadRequest)
		return models.Teacher{}, err
	} else if err != nil {
		http.Error(w, "Unable to retrive the data", http.StatusInternalServerError)
		return models.Teacher{}, err
	}
	updatedTeacher.ID = existingTeachers.ID
	res, err := db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, updatedTeacher.ID)

	if err != nil {
		http.Error(w, "Uable to update the db", http.StatusInternalServerError)
		return models.Teacher{}, err
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Rows updated: ", rowsAffected)
	return updatedTeacher, nil

}

func PatchTeacherDBHandler(db *sql.DB, id int, err error, w http.ResponseWriter, updates map[string]interface{})  error {
	var existingTeachers models.Teacher
	row := db.QueryRow(`SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?`, id)
	err = row.Scan(&existingTeachers.ID, &existingTeachers.FirstName, &existingTeachers.LastName, &existingTeachers.Email, &existingTeachers.Class, &existingTeachers.Subject)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No rows foun with this ID", http.StatusNotFound)
			return err
		}

		http.Error(w, "Unable to retrive data", http.StatusInternalServerError)
		return err
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
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Rows updated: ", rowsAffected)
	return nil
}
