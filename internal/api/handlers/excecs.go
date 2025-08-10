package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"student_management_api/Golang/internal/models"
	"student_management_api/Golang/internal/repository/sqlconnect"
	"student_management_api/Golang/pkg/utils"
	"time"
)

func ExcecsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello get request from excecs  page"))
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Something went wrong %v", err)
			return
		}
		fmt.Println("Form from POST methods =========> ", r.Form)
		w.Write([]byte("Hello post request from excecs  page"))
		fmt.Println("Hello post request from excecs  page")
	case http.MethodPut:
		w.Write([]byte("Hello put request from excecs  page"))
		fmt.Println("Hello put request from excecs  page")
	case http.MethodDelete:
		w.Write([]byte("Hello delete request from excecs  page"))
		fmt.Println("Hello delete requets from excecs  page")
	}
}

func GetExecsHndler(w http.ResponseWriter, r *http.Request) {
	var Execs []models.Exec
	Execs, err := sqlconnect.GetExecsDbHandler(Execs, r)
	if err != nil {
		return
	}
	response := struct {
		Status string
		Count  int
		Data   []models.Exec
	}{
		Status: "Success",
		Count:  len(Execs),
		Data:   Execs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetOneExecHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Println("------------ currentID", id)
	exid, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid id")
	}
	exec, err := sqlconnect.GetExecByID(exid)

	if err != nil {
		log.Println("Couldnt find the id")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exec)

}

func AddExecsHandler(w http.ResponseWriter, r *http.Request) {
	var newExecs []models.Exec
	var rawExecs []map[string]interface{} // holds the execs data before validation

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Couldn't read the incomming data", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &rawExecs)
	if err != nil {
		http.Error(w, "Couldn't parse the data", http.StatusBadRequest)
		return
	}

	fields := GetFieldNames(models.Exec{})
	allowedFields := make(map[string]struct{})
	for _, val := range fields {
		allowedFields[val] = struct{}{}
	}

	for _, exces := range rawExecs {
		for key := range exces {
			_, ok := allowedFields[key]
			if !ok {
				utils.ErrorHandler(err, "Unacceptable field")
				return
			}

		}
	}
	err = json.Unmarshal(body, &newExecs)
	if err != nil {
		http.Error(w, "Couldn't parse the incomming error", http.StatusBadRequest)
		return
	}

	for _, exces := range newExecs {
		err = CheckEmptyFields(exces)
		if err != nil {
			http.Error(w, "The is empty fiedls", http.StatusBadRequest)
			utils.ErrorHandler(err, "There is empty fields")
			return
		}
	}

	addedExecs, err := sqlconnect.AddExcecHandlerDB(newExecs)

	if err != nil {
		return
	}
	response := struct {
		Status string
		Count  int
		Data   []models.Exec
	}{
		Status: "Success",
		Count:  len(addedExecs),
		Data:   addedExecs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func PatchExecsHandler(w http.ResponseWriter, r *http.Request) {

	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.PatchExecs(updates)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

// func UpdateExecHadler(w http.ResponseWriter, r *http.Request) {
// 	idStr := r.PathValue("id")

// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		log.Println(err)
// 		fmt.Println(w, "Invalid exces ID: ", http.StatusBadRequest)
// 		return
// 	}

// 	var updatedexces models.Exec
// 	err = json.NewDecoder(r.Body).Decode(&updatedexces)
// 	if err != nil {
// 		// http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
// 		utils.ErrorHandler(err, "Invalid request payload")
// 		return
// 	}

// 	updatedexcesDB, err := sqlconnect.UpdateExecDBHandler(id, updatedexces)
// 	if err != nil {
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(updatedexcesDB)
// }

func PatchOneExecsHandler(w http.ResponseWriter, r *http.Request) {
	// idStr := strings.TrimPrefix(r.URL.Path, "/execs/")
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid exec ID", http.StatusBadRequest)
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusInternalServerError)
	}

	updatedexec, err := sqlconnect.PatchExecDBHandler(id, updates)
	if err != nil {
		log.Println(err)
		return
	}

	json.NewEncoder(w).Encode(updatedexec)

}

func DeleteExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/execs/")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		// http.Error(w, "Request not found", http.StatusBadRequest)
		utils.ErrorHandler(err, "BadRequest")
		return
	}

	affectedRow, err := sqlconnect.DeleteOneExec(w, id)
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
		Status: "exec Successfully deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteExecsHandler(w http.ResponseWriter, r *http.Request) {
	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Println("Invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	deletedIDs, err := sqlconnect.DeleteExecs(ids)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status     string `json:"status"`
		DeletedIDs []int  `json:"deleted_IDs"`
	}{
		Status:     "Execs successfully deleted",
		DeletedIDs: deletedIDs,
	}

	json.NewEncoder(w).Encode(response)

}

func LoginExecHadler(w http.ResponseWriter, r *http.Request) {
	var req models.Exec
	// Data Validation
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if req.UserName == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Search for the use if the user actually exists
	err, user, shouldReturn := sqlconnect.GetUserByUsername(w, req)
	if shouldReturn {
		return
	}
	// is user active
	if !user.InactiveStatus {
		http.Error(w, "The user is not active", http.StatusForbidden)
		utils.ErrorHandler(err, "Inactive user try to login")
		return
	}

	err = utils.VerifyPassword(req.Password, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	// Generate token
	tokenstring, err := utils.SignToken(user.ID, user.UserName, user.Role)
	if err != nil {
		http.Error(w, "Couldn't create login for the current user", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    tokenstring,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "test",
		Value:    "test string",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenstring,
	}

	json.NewEncoder(w).Encode(response)

	// send token as a response or as a cookie
}

func LogoutExecHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{ 
	message: logged out successfully 
				}`))
}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	_ = id

	var req models.UpdatePasswordModel
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	r.Body.Close()

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "Please enter password", http.StatusBadRequest)
		return
	}

	_, err = sqlconnect.UpdatePasswordFromDB(id, req.CurrentPassword, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "applocation/json")
	reponse := struct {
		Message string `json:"message"`
	}{
		Message: "Password updated",
	}

	json.NewEncoder(w).Encode(reponse)

}
