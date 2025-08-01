package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"student_management_api/Golang/internal/models"
	"student_management_api/Golang/pkg/utils"
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
		// http.Error(w, , http.StatusBadRequest)
		return nil, utils.ErrorHandler(err, fmt.Sprintf("Couldn't connect to db: %v", err))
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
		return nil, utils.ErrorHandler(err, "Databse Query Error")
	}
	defer rows.Close()

	teachersList := make([]models.Teacher, 0)

	for rows.Next() {
		var teacher models.Teacher
		err = rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			return nil, utils.ErrorHandler(err,"Error scanning databse results")
		}
		teachersList = append(teachersList, teacher)
		fmt.Println(teachersList)
	}
	return teachersList, nil
}

func GetTeacherById(id int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		utils.ErrorHandler(err, fmt.Sprintf("Couldn't connect to db: %v", err))
	}
	defer db.Close()

	var teacher models.Teacher
	err = db.QueryRow(`SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?`, id).Scan(
		&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)

	if err == sql.ErrNoRows {
		return models.Teacher{}, utils.ErrorHandler(err, "No rows found with this ID")
	} else if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, fmt.Sprintf("Databse query error: %v", err))
	}
	return teacher, nil
}

func AddTeachersDBHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()

	if err != nil {
		return nil, utils.ErrorHandler(err, "Couldn't connect to the database")
	}

	defer db.Close()

	// stmt, err := db.Prepare(`INSERT INTO teachers(first_name, last_name, email, class, subject) VALUES(?,?,?,?,?)`)
	stmt, err := db.Prepare(generateInsertQuery(models.Teacher{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, fmt.Sprintf("Database preparation failed: %v", err))
	}

	defer stmt.Close()

	addTeacers := make([]models.Teacher, len(newTeachers))

	for i, teacher := range newTeachers {
		// res, err := stmt.Exec(teacher.FirstName, teacher.LastName, teacher.Email, teacher.Class, teacher.Subject)
		values := getStructValues(teacher)
		res, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid request")
		}
		id, err := res.LastInsertId()
		if err != nil {
			utils.ErrorHandler(err, "Couldnt fetch the ID")
		}
		teacher.ID = int(id)
		addTeacers[i] = teacher

	}
	return addTeacers, err
}

func generateInsertQuery(model interface{}) string {
	modelType := reflect.TypeOf(model)
	var columns, placeholders string
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		fmt.Println(dbTag)
		dbTag = strings.TrimSuffix(dbTag, ",omitempty")
		if dbTag != "" && dbTag != "id" {
			if columns != "" {
				columns += ", "
				placeholders += ", "
			}
			columns += dbTag
			placeholders += "?"
		}
	}
	return fmt.Sprintf("INSERT INTO teachers (%s) VALUES (%s)", columns, placeholders)
}

func getStructValues(model interface{}) []interface{} {
	modelValue := reflect.ValueOf(model)
	modelType := modelValue.Type()
	values := []interface{}{}
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		if dbTag != "" && dbTag !=  "id,omitempty" {
			values = append(values, modelValue.Field(i).Interface())
		}
		
	}
	log.Println("Values: ", values)
	return values
}


func UpdateTeacherDBHandler(id int, updatedTeacher models.Teacher) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Couldn't connect to the database")
		// http.Error(w, "Unable to connect", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Unable to connect")
	}

	defer db.Close()

	var existingTeachers models.Teacher
	row := db.QueryRow(`SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?`, id)

	err = row.Scan(&existingTeachers.ID, &existingTeachers.FirstName, &existingTeachers.LastName, &existingTeachers.Email, &existingTeachers.Class, &existingTeachers.Subject)

	if err == sql.ErrNoRows {
		return models.Teacher{}, utils.ErrorHandler(err, "No row's found with this id")
	} else if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Unable to retrive the data")
	}
	updatedTeacher.ID = existingTeachers.ID
	res, err := db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, updatedTeacher.ID)

	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Uable to update the db")
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Rows updated: ", rowsAffected)
	return updatedTeacher, nil

}

func PatchTeacherDBHandler(id int, updates map[string]interface{}) (models.Teacher, error) {

	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Couldn't connect to thw database")
	}

	defer db.Close()

	var existingTeacher models.Teacher
	row := db.QueryRow(`SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?`, id)
	err = row.Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err != nil {
		if err == sql.ErrNoRows {
			// http.Error(w, "No rows foun with this ID", http.StatusNotFound)
			return models.Teacher{}, utils.ErrorHandler(err, "No rows foun with this ID")
		}

		// http.Error(w, "Unable to retrive data", http.StatusInternalServerError)
		return models.Teacher{}, err
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

	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
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

	res, err := db.Exec(`UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?`, existingTeacher.FirstName, existingTeacher.LastName, existingTeacher.Email, existingTeacher.Class, existingTeacher.Subject, existingTeacher.ID)
	if err != nil {
		// http.Error(w, fmt.Sprintf("Couldn't update the reuested data: %v", err), http.StatusNotFound)
		return models.Teacher{}, utils.ErrorHandler(err, "Couldn't update the reuested data")
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Rows updated: ", rowsAffected)
	return existingTeacher, nil
}

func DeleteOneTeacher(w http.ResponseWriter, id int) (int64, error) {
	db, err := ConnectDB()
	if err != nil {
		return 0, utils.ErrorHandler(err,  "Couldn't connect to db")
	}

	defer db.Close()

	res, err := db.Exec(`DELETE FROM teachers WHERE id = ?`, id)

	if err != nil {
		return 0, utils.ErrorHandler(err, "No request data found")
	}

	affectedRow, err := res.RowsAffected()
	if err != nil {
		fmt.Println("No affected row is found")
		return 0, utils.ErrorHandler(err, "No affected row is found")
	}
	return affectedRow, err
}

func DeleteTeachers(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Couldn't connect to db")
	}

	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Println("Invlid transaction")
		return nil, utils.ErrorHandler(err, "Invalid transaction")
	}

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Invalid transaction")
	}

	defer stmt.Close()

	deletedIDs := []int{}

	for _, id := range ids {
		res, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			log.Println("Invlid transaction")
			return nil, utils.ErrorHandler(err, "Invalid transaction")
		}

		affectedRow, err := res.RowsAffected()
		if err != nil {
			log.Println("Invlid transaction")
			return nil, utils.ErrorHandler(err, "Invalid transaction")
		}

		if affectedRow > 0 {
			deletedIDs = append(deletedIDs, id)
		}

		if affectedRow < 1 {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Invalid transaction")
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("err")
		return nil, utils.ErrorHandler(err, "Invalid transaction")
	}

	if len(deletedIDs) < 1 {
		return nil, utils.ErrorHandler(err, "IDs do not exists")
	}
	return deletedIDs, nil
}
