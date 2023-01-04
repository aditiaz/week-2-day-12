package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"my-web/connection"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type MetaData struct {
	Title     string
	IsLogin   bool
	UserName  string
	FlashData string
}

var Data = MetaData{
	Title: "Personal Web",
}

type dataInput struct{
	Id int
	ProjectName string
	Description string
	Technologies []string
	start_date time.Time
	end_date time.Time
	Duration string
	Image string
	IsLogin     bool
} 

type User struct {
	Id       int
	Email    string
	Name     string
	Password string
}

var dataInputs = []dataInput{
	{
	
	
		
	},
}


func main() {
	route := mux.NewRouter()
	connection.DataBaseConnection()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/",http.FileServer(http.Dir("./public"))))
	
	route.HandleFunc("/home",home).Methods("GET")
	route.HandleFunc("/home",addMyProject).Methods("POST")
	route.HandleFunc("/editProject/{id}",editProject).Methods("GET")
	route.HandleFunc("/updateProject/{id}",updateProject).Methods("POST")
	route.HandleFunc("/projectDetail/{id}",projectDetail).Methods("GET")
	route.HandleFunc("/contactMe",contactMe).Methods("GET")
	route.HandleFunc("/addProject",addProject).Methods("GET")
	route.HandleFunc("/delete-Project/{id}", deleteProject).Methods("GET")

	route.HandleFunc("/register", formRegister).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")

	route.HandleFunc("/login", formLogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")

	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("Server is running on port 5000")
	http.ListenAndServe("localhost:5000", route)
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/register.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	name := r.PostForm.Get("name")
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user( email,username, password) VALUES ($1, $2, $3)",  email,name, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	session.AddFlash("Successfully register!", "message")

	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	// cookie = storing data
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")
// dia menerima var flash yg namanya  "message"
	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func login(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).Scan(
		&user.Id,
		&user.Email,
		&user.Name, 
		&user.Password,
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	// not from struct :
	session.Values["IsLogin"] = true
	// not from struct :
	session.Values["Name"] = user.Name
	session.Options.MaxAge = 10800 // 1 jam = 3600 detik | 3 jam = 10800

	session.AddFlash("successfully login!", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/home", http.StatusMovedPermanently)
}


func logout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("logout.")
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")
	session.Options.MaxAge = -1 // gak boleh kurang dari 0
	session.Save(r, w)

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}




func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}
	var result []dataInput
	rows, err := connection.Conn.Query(context.Background(), "SELECT id, name, description, technologies, start_date, end_date FROM tb_projects ")
	if err != nil {
		fmt.Println(err.Error())
		return 
	}

	for rows.Next() {
		var each = dataInput{}

		var err = rows.Scan(&each.Id, &each.ProjectName,&each.Description,&each.Technologies,&each.start_date, &each.end_date)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
       each.Duration = period(each.start_date, each.end_date)
		result = append(result, each)
	}
	fmt.Println(result)

	// auth
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}
	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)

		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}
	Data.FlashData = strings.Join(flashes, "")



	respData := map[string]interface{}{
		"Data":  Data,
		"dataInputs": result,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w,respData)
}

func projectDetail( w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html;charset=utf-8")

	id,_ := strconv.Atoi(mux.Vars(r)["id"])
	var tmpl,err = template.ParseFiles("views/detail-page.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}
	
	ProjectDetail := dataInput{}
	err =  connection.Conn.QueryRow(context.Background(),"SELECT id, name, description, technologies, start_date, end_date FROM tb_projects WHERE id=$1",id).Scan(
		&ProjectDetail.Id, &ProjectDetail.ProjectName,&ProjectDetail.Description,&ProjectDetail.Technologies,&ProjectDetail.start_date, &ProjectDetail.end_date)
	    if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("message : " + err.Error()))
			return
		}
		ProjectDetail.Duration = period(ProjectDetail.start_date, ProjectDetail.end_date)
		Start_date :=   ProjectDetail.start_date.Format("02-Jan-2006")
		End_date :=   ProjectDetail.end_date.Format("02-Jan-2006")

		// auth :
	
	

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}
	resp := map[string]interface{}{
		"dataInputs" : ProjectDetail,
		"start_date" : Start_date,
		"end_date" : End_date,
	    "Data"    : Data,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w,resp,)

}

func period(start time.Time, end time.Time)string{

	distance := end.Sub(start)

	var duration string
	year := int(distance.Hours()/(12 * 30 * 24))
	 if year != 0 {
		duration = strconv.Itoa(year) + " tahun"
	}else{
		month := int(distance.Hours()/(30 * 24))
		if month != 0 {
			duration = strconv.Itoa(month) + " bulan"
		}else{
			week := int(distance.Hours()/(7 *24))
			if week != 0 {
				duration = strconv.Itoa(week) +  " minggu"
			} else {
				day := int(distance.Hours()/(24))
				if day != 0 {
					duration = strconv.Itoa(day) + " hari"
				}
			}
		}
	}	
	return duration
}






func addMyProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}  
    projectName :=  r.PostForm.Get("name")
	description := r.PostForm.Get("description")
	checkbox := r.Form["checkbox"]
	startDate := r.PostForm.Get("start-date")
	endDate := r.PostForm.Get("end-date")
	startDateTime, _ := time.Parse("2006-01-02", startDate)

	// End Date
	endDateTime, _ := time.Parse("2006-01-02", endDate)
	_,err = connection.Conn.Exec(context.Background(),"INSERT INTO tb_projects(name, description, technologies, start_date, end_date) VALUES ($1,$2,$3,$4,$5)",projectName,description,checkbox,startDateTime,endDateTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	// fmt.Println(dataInputs)
	http.Redirect(w,r,"/home",http.StatusMovedPermanently)
}



func updateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}  
    projectName :=  r.PostForm.Get("name")
	startDate := r.PostForm.Get("start-date")
	endDate := r.PostForm.Get("end-date")
	description := r.PostForm.Get("description")
	checkbox := r.Form["checkbox"]
	// image := r.PostForm.Get("image")
	Start_date,_ :=   time.Parse("2006-01-02",startDate)
	End_date,_ :=   time.Parse("2006-01-02",endDate)

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_projects SET name = $1, description = $2, technologies = $3, start_date = $4, end_date = $5 WHERE id=$6", projectName, description, checkbox, Start_date, End_date, id)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("message : " + err.Error()))
        return
    }
	
		
	http.Redirect(w,r,"/home",http.StatusMovedPermanently)
}

func editProject( w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
    var tmpl,err = template.ParseFiles("views/update.html")
	id,_ := strconv.Atoi(mux.Vars(r)["id"])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}
	
	ProjectDetail := dataInput{}
	err =  connection.Conn.QueryRow(context.Background(),"SELECT id, name, description, technologies, start_date, end_date FROM tb_projects WHERE id=$1",id).Scan(
		&ProjectDetail.Id, &ProjectDetail.ProjectName,&ProjectDetail.Description,&ProjectDetail.Technologies,&ProjectDetail.start_date, &ProjectDetail.end_date)
	    if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("message : " + err.Error()))
			return
		}
		Start_date :=   ProjectDetail.start_date.Format("01-02-2006")
		End_date :=   ProjectDetail.end_date.Format("01-02-2006")

		var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	    session, _ := store.Get(r, "SESSION_ID")

	   if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	   } else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	  }
	resp := map[string]interface{}{
		"dataInputs" : ProjectDetail,
		"start_date" : Start_date,
		"end_date" : End_date,
		"Data" : Data,
	}
	
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w,resp)
}





func deleteProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	_,err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id=$1", id)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("message : " + err.Error()))
        return
	}

	http.Redirect(w, r, "/home", http.StatusFound)
}

func contactMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/get-in-touch.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w,Data)
}
func addProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/add-project.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w,Data)
}
