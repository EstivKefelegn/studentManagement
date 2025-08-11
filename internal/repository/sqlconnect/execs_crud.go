package sqlconnect

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"student_management_api/Golang/internal/models"
	"student_management_api/Golang/pkg/utils"
	"time"

	"github.com/go-mail/mail/v2"
)

func GetExecsDbHandler(exces []models.Exec, r *http.Request) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {

		return nil, utils.ErrorHandler(err, "Couldn't connect to the database")
	}

	defer db.Close()
	query := "SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE 1=1"

	var args []interface{}
	query, args = utils.AddFilter(r, query, args)

	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		utils.ErrorHandler(err, "Can't query the rows")
		return nil, utils.ErrorHandler(err, "Can't query the rows")
	}

	defer rows.Close()

	execList := make([]models.Exec, 0)

	for rows.Next() {
		var exec models.Exec
		err = rows.Scan(&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email, &exec.UserName, &exec.UserCreatedAt, &exec.InactiveStatus, &exec.Role)

		if err != nil {
			utils.ErrorHandler(err, "unknown error")
			return nil, utils.ErrorHandler(err, "unknown error")
		}
		execList = append(execList, exec)
	}
	return execList, nil
}

func GetExecByID(execID int) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		utils.ErrorHandler(err, "Couldn't connect to the database")
	}

	defer db.Close()
	fmt.Println("Exec ID is ========= ", execID)
	var exec models.Exec
	err = db.QueryRow(`SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE id = ?`, execID).Scan(
		&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email, &exec.UserName, &exec.UserCreatedAt, &exec.InactiveStatus, &exec.Role,
	)
	fmt.Println("Exec Value is ========= ", "ID", exec.ID, exec.FirstName, exec.LastName)

	if err == sql.ErrNoRows {
		return models.Exec{}, utils.ErrorHandler(err, "No rows found with this ID")
	} else if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, fmt.Sprintf("Databse query error: %v", err))
	}

	return exec, nil
}

func AddExcecHandlerDB(newExces []models.Exec) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return []models.Exec{}, utils.ErrorHandler(err, "Couldn't connect to thedatbase")
	}

	defer db.Close()

	stmt, err := db.Prepare(utils.GenerateInsertQuery("execs", models.Exec{}))

	fmt.Println("----------------", stmt)
	if err != nil {
		return []models.Exec{}, utils.ErrorHandler(err, "Database preparation failed")
	}

	defer stmt.Close()

	addExecs := make([]models.Exec, len(newExces))

	for i, exec := range newExces {
		exec.Password, err = utils.HashPassword(exec.Password)
		if err != nil {
			return nil, err
		}

		values := utils.GetStructValues(exec)
		res, err := stmt.Exec(values...)
		if err != nil {
			return []models.Exec{}, utils.ErrorHandler(err, "Invalid ID")
		}
		id, err := res.LastInsertId()
		if err != nil {
			return []models.Exec{}, utils.ErrorHandler(err, "Couldn't fetch the id")
		}

		exec.ID = int(id)
		addExecs[i] = exec
	}

	return addExecs, nil
}

func UpdateExecDBHandler(id int, updatedexec models.Exec) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println("Couldn't connect to the database")
		// http.Error(w, "Unable to connect", http.StatusInternalServerError)
		return models.Exec{}, utils.ErrorHandler(err, "Unable to connect")
	}

	defer db.Close()
	fmt.Println("The current id is =======> ", id)
	var existingexecs models.Exec
	row := db.QueryRow(`SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE id = ?`, id)

	err = row.Scan(&existingexecs.ID, &existingexecs.FirstName, &existingexecs.LastName, &existingexecs.Email, &existingexecs.UserName, &existingexecs.UserCreatedAt, &existingexecs.InactiveStatus, &existingexecs.Role)
	fmt.Println("The current value is: ", "ID: ", existingexecs.ID, existingexecs.FirstName, existingexecs.LastName)

	if err == sql.ErrNoRows {
		return models.Exec{}, utils.ErrorHandler(err, "No row's found with this id")
	} else if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Unable to retrive the data")
	}
	updatedexec.ID = existingexecs.ID
	res, err := db.Exec("UPDATE execs SET first_name = ?, last_name = ?, email = ?, username = ?, user_created_at = ?, inactive_status = ?, role = ? WHERE id = ?", updatedexec.FirstName, updatedexec.LastName, updatedexec.Email, updatedexec.UserName, updatedexec.UserCreatedAt, updatedexec.Role, updatedexec.ID)

	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Uable to update the db")
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Rows updated: ", rowsAffected)
	return updatedexec, nil

}

func PatchExecDBHandler(id int, updates map[string]interface{}) (models.Exec, error) {

	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Couldn't connect to thw database")
	}

	defer db.Close()

	var existingexec models.Exec
	row := db.QueryRow(`SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE id = ?`, id)
	err = row.Scan(&existingexec.ID, &existingexec.FirstName, &existingexec.LastName, &existingexec.Email, &existingexec.UserName, &existingexec.UserCreatedAt, &existingexec.InactiveStatus, &existingexec.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			// http.Error(w, "No rows foun with this ID", http.StatusNotFound)
			return models.Exec{}, utils.ErrorHandler(err, "No rows foun with this ID")
		}

		// http.Error(w, "Unable to retrive data", http.StatusInternalServerError)
		return models.Exec{}, err
	}

	// for k, v := range updates {
	// 	switch k {
	// 	case "first_name":
	// 		existingexecs.FirstName = v.(string)
	// 	case "last_name":
	// 		existingexecs.LastName = v.(string)
	// 	case "email":
	// 		existingexecs.Email = v.(string)
	// 	case "subject":
	// 		existingexecs.Subject = v.(string)

	// 	}
	// }

	execVal := reflect.ValueOf(&existingexec).Elem()
	execType := execVal.Type()

	for k, v := range updates {
		for i := 0; i < execVal.NumField(); i++ {
			field := execType.Field(i)
			fmt.Println("The current field is: ", field)

			if field.Tag.Get("json") == k+",omitempty" {
				if execVal.Field(i).CanSet() {
					fieldVal := execVal.Field(i)
					fieldVal.Set(reflect.ValueOf(v).Convert(execVal.Field(i).Type()))
				}
			}

		}
	}
	fmt.Println(execVal)

	res, err := db.Exec(`UPDATE execs SET 
									first_name = ?, 
									last_name = ?,
									email = ?, 
									username = ?, 
									user_created_at = ?, 
									inactive_status = ?, 
									role = ?  
									WHERE id = ?`,
		existingexec.FirstName,
		existingexec.LastName,
		existingexec.Email,
		existingexec.UserName,
		existingexec.UserCreatedAt,
		existingexec.InactiveStatus,
		existingexec.Role,
		existingexec.ID,
	)
	if err != nil {
		// http.Error(w, fmt.Sprintf("Couldn't update the reuested data: %v", err), http.StatusNotFound)
		return models.Exec{}, utils.ErrorHandler(err, "Couldn't update the reuested data")
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Rows updated: ", rowsAffected)
	return existingexec, nil
}

func DeleteOneExec(w http.ResponseWriter, id int) (int64, error) {
	db, err := ConnectDB()
	if err != nil {
		return 0, utils.ErrorHandler(err, "Couldn't connect to db")
	}

	defer db.Close()

	res, err := db.Exec(`DELETE FROM execs WHERE id = ?`, id)

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

func DeleteExecs(ids []int) ([]int, error) {
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

	stmt, err := tx.Prepare("DELETE FROM execs WHERE id = ?")
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

func PatchExecs(updates []map[string]interface{}) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			return utils.ErrorHandler(err, "invalid Id")
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "invalid Id")
		}

		var ExecFromDb models.Exec
		err = db.QueryRow("SELECT id, first_name, last_name, email, username FROM execs WHERE id = ?", id).Scan(&ExecFromDb.ID, &ExecFromDb.FirstName, &ExecFromDb.LastName, &ExecFromDb.Email, &ExecFromDb.UserName)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return utils.ErrorHandler(err, "Exec not found")
			}
			return utils.ErrorHandler(err, "error updating data")
		}

		execVal := reflect.ValueOf(&ExecFromDb).Elem()
		execType := execVal.Type()

		for k, v := range update {
			if k == "id" {
				continue // skip updating the ID field
			}
			for i := 0; i < execVal.NumField(); i++ {
				field := execType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := execVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(v)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							log.Printf("cannot convert %v to %v", val.Type(), fieldVal.Type())
							return utils.ErrorHandler(err, "error updating data")
						}
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE execs SET first_name = ?, last_name = ?, email = ?, username = ? WHERE id = ?", ExecFromDb.FirstName, ExecFromDb.LastName, ExecFromDb.Email, ExecFromDb.UserName, ExecFromDb.ID)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "error updating data")
		}
	}

	err = tx.Commit()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	return nil
}

func GetUserByUsername(w http.ResponseWriter, req models.Exec) (error, *models.Exec, bool) {
	db, err := ConnectDB()
	if err != nil {
		utils.ErrorHandler(err, "Couldn't connect to the database")
		http.Error(w, "Couldn't connect to the db", http.StatusInternalServerError)
		return nil, nil, true
	}

	defer db.Close()

	user := &models.Exec{}
	err = db.QueryRow(`SELECT id, first_name, last_name, email, username, password, inactive_status, role FROM execs WHERE username = ?`, req.UserName).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.UserName, &user.Password, &user.InactiveStatus, &user.Role,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.ErrorHandler(err, "user not found")
			http.Error(w, "user not found", http.StatusBadRequest)
			return nil, nil, true
		}
		http.Error(w, "database query error", http.StatusInternalServerError)
		return nil, nil, true

	}
	return err, user, false
}

func UpdatePasswordFromDB(id int, currentPassword string, newPassword string) (bool, error) {
	db, err := ConnectDB()
	if err != nil {
		utils.ErrorHandler(err, "couldn't connect to the db")
		return false, err
	}

	defer db.Close()

	var username string
	var userPassword string
	var userRole string
	err = db.QueryRow(`SELECT username, password, role FROM execs WHERE id = ?`, id).Scan(&username, &userPassword, &userRole)
	if err != nil {
		utils.ErrorHandler(err, "User not found")
		return false, err
	}

	err = utils.VerifyPassword(currentPassword, userPassword)
	if err != nil {
		utils.ErrorHandler(err, "The password you entered doesn't match the current password on file")
		return false, err
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		utils.ErrorHandler(err, "internal error")
		return false, err
	}

	currentTime := time.Now().Format(time.RFC3339)
	_, err = db.Exec("UPDATE execs SET password = ?, password_changed_at = ? WHERE id = ?", hashedPassword, currentTime, id)
	if err != nil {
		utils.ErrorHandler(err, "password update failed")
		return false, err
	}

	// If we want to generate a new token to update password we can uncomment
	// token, err := utils.SignToken(id, username, userRole)
	// if err != nil {
	// 	utils.ErrorHandler(err, "Upassword update couldn't create the token")
	// 	return false, err
	// }

	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "Bearer",
	// 	Value:    token,
	// 	Path:     "/",
	// 	HttpOnly: true,
	// 	Secure:   true,
	// 	Expires:  time.Now().Add(24 * time.Minute),
	// 	SameSite: http.SameSiteStrictMode,
	// })

	return true, nil
}

func ForgotPasswordDBHandler(emailID string) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Couldn't connect to the database")
	}

	defer db.Close()

	var exec models.Exec
	err = db.QueryRow(`SELECT id FROM execs WHERE email = ?`, emailID).Scan(&exec.ID)
	if err != nil {
		return utils.ErrorHandler(err, "User not found")

	}

	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset mail")
	}

	mins := time.Duration(duration)
	expiryDate := time.Now().Add(mins * time.Minute).Format(time.RFC3339)
	tokenByte := make([]byte, 32)

	_, err = rand.Read(tokenByte)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset mail")
	}

	log.Println("tokenBytes: ", tokenByte)
	token := hex.EncodeToString(tokenByte)
	log.Println("token: ", token)

	hashedToken := sha256.Sum256(tokenByte)
	log.Println("Hashed Token", hashedToken)

	hashedTokenString := hex.EncodeToString(hashedToken[:])
	fmt.Println("Token hash string: ", hashedTokenString)
	_, err = db.Exec("UPDATE execs SET password_reset_token = ?, password_token_expires = ? WHERE id = ?", hashedTokenString, expiryDate, exec.ID)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset mail")
	}

	resetURL := fmt.Sprintf("https://localhost:3000/execs/resetpassword/reset/%s", token)
	message := fmt.Sprintf("Foregot your password? Reset your password using the following link: \n%s\nIf you didn't request a password	reset, please ingnor this email. This link is only valid for %d minutes", resetURL, int(mins))

	m := mail.NewMessage()
	m.SetHeader("From", "Estifanoakefelegn1991@gmail.com")
	m.SetHeader("To", emailID)
	m.SetHeader("Subject", "Your password reset link")
	m.SetBody("text/plain", message)

	d := mail.NewDialer("localhost", 1025, "", "")
	err = d.DialAndSend(m)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset mail")
	}
	return nil
}

func ResetPasswordDBHandler(token string, newPassword string) error {
	bytes, err := hex.DecodeString(token)
	if err != nil {
		return utils.ErrorHandler(err, "internal error")
		
	}

	hashToken := sha256.Sum256(bytes)
	hashedTokenString := hex.EncodeToString(hashToken[:])

	db, err := ConnectDB()
	if err != nil {
		// http.Error(w, "Couldn't connect to the db", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "internal error")
		
	}	

	defer db.Close()

	var user models.Exec

	query := "SELECT id, email FROM execs WHERE password_reset_token = ? AND password_token_expires > ?"
	err = db.QueryRow(query, hashedTokenString, time.Now().Format(time.RFC3339)).Scan(&user.ID, &user.Email)
	if err != nil {
		return utils.ErrorHandler(err, "invalid or expired reset code")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return	utils.ErrorHandler(err, "internal error")
	}

	updateQuery := "UPDATE execs SET password  = ?, password_reset_token = NULL, password_token_expires = NULL, password_changed_at = ? WHERE id = ?"
	_, err = db.Exec(updateQuery, hashedPassword, time.Now().Format(time.RFC3339), user.ID)

	if err != nil {
		return utils.ErrorHandler(err, "Internal error")
	}
	return nil
}
