package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

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
		addr = "127.0.0.1:8000"
	}

	log.Fatal(http.ListenAndServe(addr, a.Router))
}

// initializeRoutes - creates routers, runs automatically in Initialize
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/login", a.login).Methods("POST")
	a.Router.HandleFunc("/login", a.loginOptions).Methods("OPTIONS")
	a.Router.HandleFunc("/register", a.register).Methods("POST")
	a.Router.HandleFunc("/register", a.registerOptions).Methods("OPTIONS")
	a.Router.HandleFunc("/profile", a.profile).Methods("GET")
	a.Router.HandleFunc("/profile", a.profileOptions).Methods("OPTIONS")
}

// login function return token in success
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var u User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&u)

	if err != nil {
		log.Fatalf("cannot decode signup body: %v", err)
	}

	// do validation by funcValidator
	emailValidator := v.FromFunc(func(field v.Field) v.Errors {
		val := field.ValuePtr.(*string)
		matched, err := regexp.MatchString("(^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$)", *val)
		if err != nil {
			return v.NewErrors("email", v.ErrInvalid, fmt.Sprint(err))
		}
		if matched == false {
			return v.NewErrors("email", v.ErrInvalid, "Wrong email")
		}
		return nil
	})

	errs := v.Validate(v.Schema{
		v.F("email", &u.Email):       v.All(v.Nonzero("cannot be empty"), v.Len(4, 120, "length is not between 4 and 120"), emailValidator),
		v.F("password", &u.Password): v.All(v.Nonzero("cannot be empty"), v.Len(4, 120, "length is not between 8 and 120")),
	})

	if len(errs) > 0 {
		respondWithJSON(w, r, http.StatusBadRequest, errs.JSONErrors())
		return
	}

	if err := a.Storage.GetUserByEmail(&u); err != nil {
		errs.Append(v.NewError("__error__", v.ErrInvalid, "email or password do not match"))
	}

	if len(errs) > 0 {
		respondWithJSON(w, r, http.StatusBadRequest, errs.JSONErrors())
		return
	}

	t, err := u.GetToken()
	if err != nil {
		errs.Append(v.NewError("__error__", v.ErrInvalid, fmt.Sprintf("%v", err)))
		respondWithJSON(w, r, http.StatusBadRequest, errs.JSONErrors())
		return
	}

	respondWithJSON(w, r, http.StatusOK, map[string]string{"token": t})
}

// login options request
// usefull to have same validation rules on front and back end
func (a *App) loginOptions(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, r, 200, map[string]map[string]string{
		"email":    map[string]string{"type": "string", "required": "1", "maxLength": "255"},
		"password": map[string]string{"type": "password", "required": "1", "maxLength": "255"},
	})
}

// register function, return jwt token in success
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

	// do validation by funcValidator
	emailValidator := v.FromFunc(func(field v.Field) v.Errors {
		val := field.ValuePtr.(*string)
		matched, err := regexp.MatchString("(^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$)", *val)
		if err != nil {
			return v.NewErrors("email", v.ErrInvalid, fmt.Sprint(err))
		}
		if matched == false {
			return v.NewErrors("email", v.ErrInvalid, "Wrong email")
		}
		return nil
	})

	errs := v.Validate(v.Schema{
		v.F("email", &u.Email):       v.All(v.Nonzero("cannot be empty"), v.Len(4, 120, "length is not between 4 and 120"), emailValidator),
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
		respondWithJSON(w, r, http.StatusCreated, res)
	}
}

// register options function - for frontend validation rules
func (a *App) registerOptions(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, r, 200, map[string]map[string]string{
		"email":    map[string]string{"type": "string", "required": "1", "minLength": "4", "maxLength": "255"},
		"password": map[string]string{"type": "password", "required": "1", "minLength": "8", "maxLength": "255"},
	})
}

// profile function, return user data in success
func (a *App) profile(w http.ResponseWriter, r *http.Request) {

	u := User{}
	token := r.Header.Get("Authorization")
	if token == "" {
		respondWithError(w, r, http.StatusUnauthorized, "Authorization")
		return
	}

	res, err := u.InvalidToken(token)
	if err != nil {
		respondWithError(w, r, http.StatusForbidden, fmt.Sprintf("%v", err))
		return
	}

	if res == false {
		respondWithError(w, r, http.StatusForbidden, "invalid token")
		return
	}

	respondWithJSON(w, r, http.StatusOK, map[string]interface{}{"email": u.Email, "first_name": "", "last_name": ""})
	return
}

// register options function - for frontend validation rules
func (a *App) profileOptions(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, r, 200, map[string]map[string]string{
		"email":    map[string]string{"type": "string", "required": "1", "minLength": "4", "maxLength": "255"},
		"password": map[string]string{"type": "password", "required": "1", "minLength": "8", "maxLength": "255"},
	})
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

	/*
		 FixMe: Will stop real options requests either

		 Stop here if its Preflighted OPTIONS request
		if r.Method == "OPTIONS" && r.Header.Get("Accept") == "* / *" {
			return
		}
	*/
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Max-Age", "1200")
	}

	respondWithBytes(w, code, response)
}

// respondWithBytes simple return code and byte data
func respondWithBytes(w http.ResponseWriter, code int, response []byte) {
	w.WriteHeader(code)
	w.Write(response)
}
