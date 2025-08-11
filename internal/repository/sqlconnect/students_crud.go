package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"student_management_api/Golang/internal/models"
	"student_management_api/Golang/pkg/utils"
)

func GetStudentsDbHandler(students []models.Student, r *http.Request, page, limit int) ([]models.Student, int, error) {
	db, err := ConnectDB()

	if err != nil {
		// http.Error(w, , http.StatusBadRequest)
		return nil, 0, utils.ErrorHandler(err, fmt.Sprintf("Couldn't connect to db: %v", err))
	}
	defer db.Close()

	query := `SELECT id, first_name, last_name, email, class FROM students WHERE 1=1`
	var args []interface{}

	query, args = utils.AddFilter(r, query, args)
	
	//Add Pagination formula to add pagination is (page - 1) * limit
	offset := (page - 1) * limit
	query += " LIMIT ? OFFSET ?" 

	args = append(args, limit, offset)

	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)

	if err != nil {
		fmt.Println(err)
		// http.Error(w, "Databse Query Error", http.StatusInternalServerError)
		return nil,  0, utils.ErrorHandler(err, "Databse Query Error")
	}
	defer rows.Close()

	studentsList := make([]models.Student, 0)

	for rows.Next() {
		var student models.Student
		err = rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return nil, 0, utils.ErrorHandler(err, "Error scanning databse results")
		}
		studentsList = append(studentsList, student)
		fmt.Println(studentsList)
	}

	var totalStudents int
	err = db.QueryRow("SELECT COUNT(*) FROM students").Scan(&totalStudents)
	if err != nil {
		utils.ErrorHandler(err, "")
		totalStudents = 0
	}


	return studentsList, totalStudents, nil
}

func GetStudentById(id int) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		utils.ErrorHandler(err, fmt.Sprintf("Couldn't connect to db: %v", err))
	}
	defer db.Close()

	var student models.Student
	err = db.QueryRow(`SELECT id, first_name, last_name, email, class FROM teachers WHERE id = ?`, id).Scan(
		&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)

	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, "No rows found with this ID")
	} else if err != nil {
		return models.Student{}, utils.ErrorHandler(err, fmt.Sprintf("Databse query error: %v", err))
	}
	return student, nil
}

func AddStudentsDBHandler(newstudents []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()

	if err != nil {
		return nil, utils.ErrorHandler(err, "Couldn't connect to the database")
	}

	defer db.Close()

	// stmt, err := db.Prepare(`INSERT INTO students(first_name, last_name, email, class) VALUES(?,?,?,?,?)`)
	stmt, err := db.Prepare(utils.GenerateInsertQuery("students", models.Student{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, fmt.Sprintf("Database preparation failed: %v", err))
	}

	defer stmt.Close()

	addStudents := make([]models.Student, len(newstudents))

	for i, student := range newstudents {
		// res, err := stmt.Exec(student.FirstName, student.LastName, student.Email, student.Class, student.Subject)
		values := utils.GetStructValues(student)
		res, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid request")
		}
		id, err := res.LastInsertId()
		if err != nil {
			utils.ErrorHandler(err, "Couldnt fetch the ID")
		}
		student.ID = int(id)
		addStudents[i] = student

	}
	return addStudents, err
}

func UpdateStudentDBHandler(id int, updatedstudent models.Student) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Couldn't connect to the database")
		// http.Error(w, "Unable to connect", http.StatusInternalServerError)
		return models.Student{}, utils.ErrorHandler(err, "Unable to connect")
	}

	defer db.Close()

	var existingstudents models.Student
	row := db.QueryRow(`SELECT id, first_name, last_name, email, class FROM students WHERE id = ?`, id)

	err = row.Scan(&existingstudents.ID, &existingstudents.FirstName, &existingstudents.LastName, &existingstudents.Email, &existingstudents.Class)

	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, "No row's found with this id")
	} else if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Unable to retrive the data")
	}
	updatedstudent.ID = existingstudents.ID
	res, err := db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?", updatedstudent.FirstName, updatedstudent.LastName, updatedstudent.Email, updatedstudent.Class, updatedstudent.ID)

	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Uable to update the db")
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Rows updated: ", rowsAffected)
	return updatedstudent, nil

}

func PatchOneStudentDBHandler(id int, updates map[string]interface{}) (models.Student, error) {

	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Couldn't connect to thw database")
	}

	defer db.Close()

	var existingstudent models.Student
	row := db.QueryRow(`SELECT id, first_name, last_name, email, class FROM students WHERE id = ?`, id)
	err = row.Scan(&existingstudent.ID, &existingstudent.FirstName, &existingstudent.LastName, &existingstudent.Email, &existingstudent.Class)
	if err != nil {
		if err == sql.ErrNoRows {
			// http.Error(w, "No rows foun with this ID", http.StatusNotFound)
			return models.Student{}, utils.ErrorHandler(err, "No rows foun with this ID")
		}

		// http.Error(w, "Unable to retrive data", http.StatusInternalServerError)
		return models.Student{}, err
	}

	// for k, v := range updates {
	// 	switch k {
	// 	case "first_name":
	// 		existingstudents.FirstName = v.(string)
	// 	case "last_name":
	// 		existingstudents.LastName = v.(string)
	// 	case "email":
	// 		existingstudents.Email = v.(string)
	// 	case "subject":
	// 		existingstudents.Subject = v.(string)

	// 	}
	// }

	studentVal := reflect.ValueOf(&existingstudent).Elem()
	studentType := studentVal.Type()

	for k, v := range updates {
		for i := 0; i < studentVal.NumField(); i++ {
			field := studentType.Field(i)
			fmt.Println("The current field is: ", field)

			if field.Tag.Get("json") == k+",omitempty" {
				if studentVal.Field(i).CanSet() {
					fieldVal := studentVal.Field(i)
					fieldVal.Set(reflect.ValueOf(v).Convert(studentVal.Field(i).Type()))
				}
			}

		}
	}
	fmt.Println(studentVal.Type())

	res, err := db.Exec(`UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? = ? WHERE id = ?`, existingstudent.FirstName, existingstudent.LastName, existingstudent.Email, existingstudent.Class, existingstudent.ID)
	if err != nil {
		// http.Error(w, fmt.Sprintf("Couldn't update the reuested data: %v", err), http.StatusNotFound)
		return models.Student{}, utils.ErrorHandler(err, "Couldn't update the reuested data")
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Rows updated: ", rowsAffected)
	return existingstudent, nil
}

func DeleteOneStudent(w http.ResponseWriter, id int) (int64, error) {
	db, err := ConnectDB()
	if err != nil {
		return 0, utils.ErrorHandler(err, "Couldn't connect to db")
	}

	defer db.Close()

	res, err := db.Exec(`DELETE FROM students WHERE id = ?`, id)

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

func DeleteStudents(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Couldn't connect to db")
	}

	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		// log.Println("Invlid transaction")
		utils.ErrorHandler(err, "Invalid Transaction")
		return nil, utils.ErrorHandler(err, "Invalid transaction")
	}

	stmt, err := tx.Prepare("DELETE FROM students WHERE id = ?")
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
