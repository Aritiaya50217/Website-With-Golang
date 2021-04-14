package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

// สร้าง stucture
type Page struct {
	// fields
	Title string
	Body  []byte
}

// เก็บไฟล์ .html ที่จำเป็นไวเในตัวแปร template
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// p ใช้ฟังก์ชัน pointer ชี้ไปที่ Page
func (p *Page) save() error {
	// p ชี้ไปที่ Title ที่อยู่ใน Page
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)

}
func loadPage(title string) (*Page, error) {
	// ตัวแปร filename เก็บค่าชื่อไฟล์
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	/*
		// เก็บชื่อไฟล์ที่ลงท้ายด้วย .html ในตัวแปร t
		t, err := template.ParseFiles(tmpl + ".html")
	*/
	// เรียกใช้ตัวแปร templates
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	// ถ้าหาไฟล์ไม่เจอ
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	/*
		// หากเกิดข้อผิดพลาดระหว่างการส่งค่า
		err = t.Execute(w, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	*/
}

// การใช้ closure
// เก็บค่า func(http.ResponseWriter, *http.Request ใน fn และ HandlerFunc เป็น type
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// เก็บค่าการตรวจสอบ URL ไว้ที่ m
		m := validPath.FindStringSubmatch(r.URL.Path)
		// ถ้า m ไม่ตรงกับ path ที่เรากำหนด หรือ m มีค่า ว่าง
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}

}

// การตรวจสอบข้อผิดพลาด validPath
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	// ตรวจสอบ URL
	m := validPath.FindStringSubmatch(r.URL.Path)
	// ถ้า m เป็นค่าว่าง
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	// The title is the second subexpression.
	return m[2], nil
}

// แสดงข้อความจากไฟล์ txt ไปยังเพจ
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	/*
		// ใส่ชื่อไฟล์ ที่เราต้องการไปยัง URL
		// len("/view/ชื่อไฟล์")
		title := r.URL.Path[len("/view/"):]
	*/
	/* เอาออกเพราะได้ใช้  getTitle from the handler functions
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	*/
	// โหลดหน้าเพจ
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	/*
		// use html/template package
		// อ่านไฟล์ view.html
		t, _ := template.ParseFiles("view.html")
		t.Execute(w, p)
	*/
	// "view" คือ ชื่อไฟล์ ไม่ต้องใส่ html เพราะเรียกใช้จาก function renderTemplate
	renderTemplate(w, "view", p)

}
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	//title := r.URL.Path[len("/edit/"):]
	/*
		title, err := getTitle(w, r)
		if err != nil {
			return
		}
	*/
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	/*
		// use html/template package
		// อ่านไฟล์ edit.html
		t, _ := template.ParseFiles("edit.html")
		t.Execute(w, p)
	*/
	renderTemplate(w, "edit", p)
}
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	//title := r.URL.Path[len("/save/"):]
	/* title, err := getTitle(w, r)
	if err != nil {
		return
	}
	*/
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	// สร้างไฟล์ชื่อ TastPage.txt ภายในหรือส่วน Body มีข้อความ this is a sample Page.
	p1 := &Page{Title: "TestPage", Body: []byte("this is a sample Page.")}
	p1.save()
	// ใช้ฟังก์ชัน loadPage เพื่อ copy ข้อความด้านใน
	p2, _ := loadPage("TestPage")
	// แสดงข้อความภายใน Body
	fmt.Println(string(p2.Body))
	// แสดงหน้าเพจจาก func viewHandler
	// can wrap the handler functions with makeHandler
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
