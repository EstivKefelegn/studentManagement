package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"student_management_api/Golang/pkg/utils"

	"github.com/microcosm-cc/bluemonday"
)

// we are going to sanitize the url, query parameters and the request body

func XSSMiddleware(next http.Handler) http.Handler {
	fmt.Println("************ Initializing XSSMiddleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Sanotize the url path
		sanitizePath, err := clean(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("Original Path:", r.URL.Path)
		fmt.Println("Sanitize Path:", sanitizePath)

		// sanitizing the qury params
		params := r.URL.Query()
		sanitizedQury := make(map[string][]string)

		for key, values := range params {
			sanitizedKey, err := clean(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var sanitizedValues []string
			for _, value := range values {
				cleanValue, err := clean(value)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				sanitizedValues = append(sanitizedValues, cleanValue.(string))
			}
			sanitizedQury[sanitizedKey.(string)] = sanitizedValues
		}

		// now replace the raw request values with the cleaned one
		r.URL.Path = sanitizePath.(string)
		r.URL.RawQuery = url.Values(sanitizedQury).Encode()

		// Sanitizing request body
		if r.Header.Get("Content-Type") == "application/json" {
			if r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					utils.ErrorHandler(err, "")
					http.Error(w, "Error reading request body", http.StatusBadRequest)
					return
				}
				bodyString := strings.TrimSpace(string(bodyBytes))
				fmt.Println("Origional Body:", bodyString)

				// Reset the request body --> r.Body is type of readcloser so the sanitized values has to be converted to readcloser type
				// NopCloser accepts io.reader and returns readcloser
				r.Body = io.NopCloser(bytes.NewReader([]byte(bodyString)))

				if len(bodyString) > 0 {
					var inputData interface{}
					err := json.NewDecoder(bytes.NewReader([]byte(bodyString))).Decode(inputData)

					if err != nil {
						http.Error(w, "Invalid JSON body=-=-=", http.StatusBadRequest)
						return
					}

					fmt.Println("Original JSON data: ", inputData)

					// Sanitize the json body
					sanotizedData, err := clean(inputData)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}

					// Marshal the sanitized data back to the body
					sanotizedBody, err := json.Marshal(sanotizedData)
					if err != nil {
						http.Error(w, utils.ErrorHandler(err, "Error sanitizing body").Error(), http.StatusBadRequest)
						return
					}
					r.Body = io.NopCloser(bytes.NewReader(sanotizedBody))
					fmt.Println("Sanitized body: ", string(sanotizedBody))
				} else {
					log.Println("Request body is empty")
				}
			} else {
				log.Println("No body in the request")
			}
		} else if r.Header.Get("Content-Type") != "" {
			log.Printf("Received request with unsupported Content-Type: %s. Expected application/json", r.Header.Get("Content-Type"))
			http.Error(w, "Unsupported Contnt-Type. Please use applicaion/json", http.StatusUnsupportedMediaType)
			return
		}

		fmt.Println("Updated URL:", r.URL.String())
		next.ServeHTTP(w, r)
		fmt.Println("Sending response from XSSMiddleware ran")
	})
}

// We accept any type of data and we clean that
func clean(data interface{}) (interface{}, error) {
	// query parmeters are key and value
	// we can also get string from path parameters
	// we may also have other type of data so we deal with interface also
	// if we have data other than we return Unsupported type

	switch v := data.(type) { // when we make type assertion v also carries type assertion and data
	case map[string]interface{}: // for key value pairs
		for k, val := range v {
			v[k] = sanitizeValue(val)
		}
		return v, nil
	case []interface{}: // for list of unknown data's
		for k, val := range v {
			v[k] = sanitizeValue(val)
		}
		return v, nil
	case string: // for the commig string values
		return sanitizeString(v), nil
	default:
		return nil, utils.ErrorHandler(fmt.Errorf("unsupported Type: %T", data), fmt.Sprintf("unsupported type %T", data))
	}

}

func sanitizeValue(data interface{}) interface{} {
	// se we may accept any type of value so we have to make a swithc case
	switch v := data.(type) { // when we make type assertion v also carries type assertion and data
	case map[string]interface{}: // for key value pairs
		for k, val := range v {
			v[k] = sanitizeValue(val) // we are calling the function again inside the function because we may have embeded map in side a map
		}
		return v
	case []interface{}: // for list of unknown data's
		for k, val := range v {
			v[k] = sanitizeValue(val) // We repeatedly call the funtion again and again until we find a string value and take it to sanitizeString
		}
		return v

	case string: // for the commig string values
		return sanitizeString(v)
	default:
		return v
	}
}

// Handling cleaning a string we use bluemonday package
func sanitizeString(value string) string {
	return bluemonday.UGCPolicy().Sanitize(value)
}
