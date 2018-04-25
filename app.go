package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx"

	"github.com/gorilla/mux"
	v "github.com/webdeveloppro/validating"
)

// App holding routers and DB connection
type App struct {
	Router  *mux.Router
	Storage Storage
}

// NewApp will create new App instance and setup storage connection
func NewApp(storage Storage) (a App, err error) {
	a = App{}
	a.Router = mux.NewRouter()
	a.initializeRoutes()
	a.Storage = storage
	return a, nil
}

// Run application on 8080 port
func (a *App) Run(addr string) {

	if addr == "" {
		addr = "8000"
	}

	log.Fatal(http.ListenAndServe(":"+addr, a.Router))
}

// initializeRoutes - creates routers, runs automatically in Initialize
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/login", a.login).Methods("POST")
	a.Router.HandleFunc("/login", a.loginOptions).Methods("OPTIONS")
	a.Router.HandleFunc("/register", a.register).Methods("POST")
	a.Router.HandleFunc("/register", a.registerOptions).Methods("OPTIONS")
	a.Router.HandleFunc("/files/{fileName}", a.uploadFile).Methods("PUT")
	a.Router.HandleFunc("/files/{fileName}", a.getFile).Methods("GET")
	a.Router.HandleFunc("/files/{fileName}", a.deleteFile).Methods("DELETE")
}

// login function return token in success
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var u User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&u)

	if err != nil {
		log.Fatalf("cannot decode signup body: %v", err)
	}

	errors := make(map[string][]string, 0)
	if u.Email == "" {
		errors["email"] = append(errors["email"], "email cannot be empty")
	}
	if u.Password == "" {
		errors["password"] = append(errors["password"], "password cannot be empty")
	}

	if len(errors) > 0 {
		respondWithJSON(w, r, http.StatusBadRequest, errors)
		return
	}

	if err := a.Storage.GetUserByEmail(&u); err != nil {
		errors["__error__"] = append(errors["__error__"], "email or password do not match")
	}

	if len(errors) > 0 {
		respondWithJSON(w, r, http.StatusBadRequest, errors)
		return
	}

	t, err := u.GetToken()
	if err != nil {
		errors["__error__"] = append(errors["__error__"], fmt.Sprintf("%v", err))
		respondWithJSON(w, r, http.StatusBadRequest, errors)
		return
	}
	res := map[string]string{"token": t}

	respondWithJSON(w, r, 200, res)
}

// login options request
// usefull to have same validation rules on front and back end
func (a *App) loginOptions(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, r, 200, map[string]map[string]string{
		"email":    map[string]string{"type": "string", "required": "1", "maxLength": "255"},
		"password": map[string]string{"type": "password", "required": "1", "maxLength": "255"},
	})
}

// register function, return 204 in success
func (a *App) register(w http.ResponseWriter, r *http.Request) {
	u := User{}

	if r.Body == nil {
		respondWithJSON(w, r, http.StatusBadRequest, []string{})
		return
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&u)

	if err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, []string{})
		return
	}

	errs := v.Validate(v.Schema{
		v.F("email", &u.Email):       v.All(v.Nonzero("cannot be empty"), v.Len(4, 120, "length is not between 4 and 120")),
		v.F("password", &u.Password): v.All(v.Nonzero("cannot be empty"), v.Len(4, 120, "length is not between 8 and 120")),
	})

	// We don't want to make database query if we already know email is not valid
	if errs.HasField("email") == false {
		err = a.Storage.GetUserByEmail(&u)
		if err == nil {
			errs.Extend(v.NewErrors("email", v.ErrInvalid, "email address already exists, do you want to reset password?"))
		} else if err != pgx.ErrNoRows {
			errs.Extend(v.NewErrors("email", v.ErrUnrecognized, err.Error()))
		}
	}

	if len(errs) > 0 {
		respondWithJSON(w, r, http.StatusBadRequest, errs.JSONErrors())
		return
	}

	if err := a.Storage.CreateUser(&u); err != nil {
		errs.Extend(v.NewErrors("__error__", v.ErrInvalid, "cannot create user, please try again in few minutes"))
		log.Fatalf("insert users errors: %+v", err)
	}

	if len(errs) > 0 {
		respondWithJSON(w, r, http.StatusBadRequest, errs.JSONErrors())
	} else {
		t, err := u.GetToken()
		if err != nil {
			errs.Extend(v.NewErrors("__errors", v.ErrInvalid, fmt.Sprintf("%v", err)))
			respondWithJSON(w, r, http.StatusBadRequest, errs.JSONErrors())
			return
		}
		res := map[string]string{"token": t}
		respondWithJSON(w, r, 201, res)
	}
}

// register options function - for frontend validation rules
func (a *App) registerOptions(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, r, 200, map[string]map[string]string{
		"email":    map[string]string{"type": "string", "required": "1", "minLength": "4", "maxLength": "255"},
		"password": map[string]string{"type": "password", "required": "1", "minLength": "8", "maxLength": "255"},
	})
}

// uploadFile function return 201 in success
func (a *App) uploadFile(w http.ResponseWriter, r *http.Request) {

	if requireToken(w, r) == false {
		return
	}

	vars := mux.Vars(r)
	Filename := vars["fileName"]
	if suspiciousFileName(Filename) {
		respondWithBytes(w, http.StatusServiceUnavailable, nil)
	}

	f, err := os.OpenFile("./content/"+Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Cannot open file, %v", err)
		respondWithJSON(w, r, http.StatusInternalServerError, "Please try again in a few minutes")
		return
	}
	defer f.Close()
	io.Copy(f, r.Body)

	w.Header().Set("Location", "/content/"+Filename)
	respondWithJSON(w, r, http.StatusCreated, []string{})
}

// getFile function return 200 and file type and file content in success
func (a *App) getFile(w http.ResponseWriter, r *http.Request) {

	if requireToken(w, r) == false {
		return
	}

	vars := mux.Vars(r)
	Filename := vars["fileName"]
	if suspiciousFileName(Filename) {
		respondWithBytes(w, http.StatusServiceUnavailable, nil)
	}

	f, err := os.Open("./content/" + Filename)
	defer f.Close()
	if err != nil {
		respondWithBytes(w, http.StatusNotFound, []byte("File not found"))
		return
	}

	FileHeader := make([]byte, 512)
	f.Read(FileHeader)
	FileContentType := http.DetectContentType(FileHeader)

	FileStat, _ := f.Stat()                            //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	w.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already so we reset the offset back to 0
	f.Seek(0, 0)
	io.Copy(w, f)
	w.WriteHeader(http.StatusAccepted)
	return
}

// deleteFile function return 204 in success
func (a *App) deleteFile(w http.ResponseWriter, r *http.Request) {

	if requireToken(w, r) == false {
		return
	}

	vars := mux.Vars(r)
	Filename := vars["fileName"]
	if suspiciousFileName(Filename) {
		respondWithBytes(w, http.StatusServiceUnavailable, nil)
	}

	err := os.Remove("./content/" + Filename)
	if err != nil {
		respondWithBytes(w, http.StatusNotFound, []byte("File not found"))
		return
	}

	respondWithBytes(w, http.StatusNoContent, nil)
	return
}

// requireToken check if request have Authorization token
func requireToken(w http.ResponseWriter, r *http.Request) bool {
	if InvalidToken(r.Header.Get("Authorization")) {
		respondWithJSON(w, r, http.StatusForbidden, map[string]string{"__error__": "Token is required"})
		return false
	}
	return true
}

// suspiciousFileName is a protection from hacks and scripts execution
// If file looks suspicious function will return true
func suspiciousFileName(Filename string) bool {
	disallowedNames := []string{"../", "`", "\n"}
	for _, d := range disallowedNames {
		if strings.Contains(d, Filename) {
			return true
		}
	}
	if len(Filename) > 255 {
		return true
	}
	return false
}

// respondWithError return error code and message
func respondWithError(w http.ResponseWriter, r *http.Request, code int, message string) {
	respondWithJSON(w, r, code, map[string]string{"error": message})
}

// respondWithJSON add all headers for SPA application and return code and data
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.Marshal(payload)

	if err != nil {
		log.Fatalf("Cannot convert data to json, %v", err)
	}

	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-REAL")
		w.Header().Set("Content-Type", "application/json")
	}

	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" && r.Header.Get("Accept") == "*/*" {
		return
	}

	respondWithBytes(w, code, response)
}

// respondWithBytes simple return code and byte data
func respondWithBytes(w http.ResponseWriter, code int, response []byte) {
	w.WriteHeader(code)
	w.Write(response)
}
