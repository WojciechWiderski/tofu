package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name  string  `json:"name"`
	Age   int     `json:"age"`
	Books []*Book `json:"books"  gorm:"many2many:user_books;"`
}

type Book struct {
	gorm.Model
	Author string  `json:"author"`
	Title  string  `json:"title"`
	Users  []*User `json:"users" gorm:"many2many:user_books;"`
}
