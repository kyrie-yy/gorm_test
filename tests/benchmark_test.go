package tests_test

import (
	"testing"

	. "gorm.io/gorm/utils/tests"
)

func BenchmarkCreate(b *testing.B) {
	user := *GetUser("bench", Config{})

	for x := 0; x < b.N; x++ {
		user.ID = 0
		DB.Create(&user)
	}
}

func BenchmarkFind(b *testing.B) {
	user := *GetUser("find", Config{})
	DB.Create(&user)

	for x := 0; x < b.N; x++ {
		DB.Find(&User{}, "id = ?", user.ID)
	}
}

func BenchmarkScan(b *testing.B) {
	user := *GetUser("find", Config{})
	DB.Create(&user)

	var u User
	for x := 0; x < b.N; x++ {
		DB.Raw("select * from users where id = ?", user.ID).Scan(&u)
	}
}

func BenchmarkUpdate(b *testing.B) {
	user := *GetUser("find", Config{})
	DB.Create(&user)

	for x := 0; x < b.N; x++ {
		DB.Model(&user).Updates(map[string]interface{}{"Age": x})
	}
}

func BenchmarkDelete(b *testing.B) {
	user := *GetUser("find", Config{})

	for x := 0; x < b.N; x++ {
		user.ID = 0
		DB.Create(&user)
		DB.Delete(&user)
	}
}
